package ingester

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/apache/arrow/go/v8/arrow"
	"github.com/apache/arrow/go/v8/arrow/memory"
	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/google/pprof/profile"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/dskit/kv"
	"github.com/grafana/dskit/ring"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	"github.com/grafana/fire/pkg/gen/ingester/v1/ingesterv1connect"
	pushv1 "github.com/grafana/fire/pkg/gen/push/v1"
	"github.com/grafana/fire/pkg/profilestore"
)

func defaultIngesterTestConfig(t testing.TB) Config {
	kvClient, err := kv.NewClient(kv.Config{Store: "inmemory"}, ring.GetCodec(), nil, log.NewNopLogger())
	require.NoError(t, err)

	cfg := Config{}
	flagext.DefaultValues(&cfg)
	cfg.LifecyclerConfig.RingConfig.KVStore.Mock = kvClient
	cfg.LifecyclerConfig.NumTokens = 1
	cfg.LifecyclerConfig.ListenPort = 0
	cfg.LifecyclerConfig.Addr = "localhost"
	cfg.LifecyclerConfig.ID = "localhost"
	cfg.LifecyclerConfig.FinalSleep = 0
	cfg.LifecyclerConfig.MinReadyDuration = 0
	return cfg
}

func defaultProfileStoreTestConfig(t testing.TB) *profilestore.Config {
	cfg := &profilestore.Config{}
	flagext.DefaultValues(cfg)

	// create data path
	dataPath, err := os.MkdirTemp("", "fire-db")
	require.NoError(t, err)
	t.Logf("created temporary data path: %s", dataPath)
	t.Cleanup(func() {
		if err := os.RemoveAll(dataPath); err != nil {
			t.Logf("remove data path failed: %v", err)
		}
	})
	cfg.DataPath = dataPath

	return cfg
}

func Test_ConnectPush(t *testing.T) {
	cfg := defaultIngesterTestConfig(t)
	logger := log.NewLogfmtLogger(os.Stdout)

	profileStore, err := profilestore.New(logger, nil, trace.NewNoopTracerProvider(), defaultProfileStoreTestConfig(t))
	require.NoError(t, err)

	mux := http.NewServeMux()
	d, err := New(cfg, log.NewLogfmtLogger(os.Stdout), nil, profileStore)
	require.NoError(t, err)

	mux.Handle(ingesterv1connect.NewIngesterServiceHandler(d))
	s := httptest.NewServer(mux)
	defer s.Close()

	client := ingesterv1connect.NewIngesterServiceClient(http.DefaultClient, s.URL)

	rawProfile := testMyHeapProfile(t)
	resp, err := client.Push(context.Background(), connect.NewRequest(&pushv1.PushRequest{
		Series: []*pushv1.RawProfileSeries{
			{
				Labels: []*commonv1.LabelPair{
					{Name: "__name__", Value: "my_own_profile"},
					{Name: "cluster", Value: "us-central1"},
				},
				Samples: []*pushv1.RawSample{
					{
						RawProfile: rawProfile,
					},
				},
			},
		},
	}))
	require.NoError(t, err)
	require.NotNil(t, resp)
	ingestedSamples := countNonZeroValues(parseRawProfile(t, bytes.NewBuffer(rawProfile)))

	profileStore.Table().Sync()
	var queriedSamples int64
	require.NoError(t, profileStore.Table().View(func(tx uint64) error {
		return profileStore.Table().Iterator(context.Background(), tx, memory.NewGoAllocator(), nil, nil, nil, func(ar arrow.Record) error {
			t.Log(ar)

			queriedSamples += ar.NumRows()

			return nil
		})
	}))

	require.Equal(t, ingestedSamples, queriedSamples, "expected to query all ingested samples")

	require.NoError(t, profileStore.Table().RotateBlock())

	require.NoError(
		t,
		profileStore.Close(),
	)
}

// This counts all sample values, where at least a single value in a sample is non-zero
func countNonZeroValues(p *profile.Profile) int64 {
	var count int64
	for _, s := range p.Sample {
		for _, v := range s.Value {
			if v != 0 {
				count++
				continue
			}
		}
	}
	return count
}

func parseRawProfile(t testing.TB, reader io.Reader) *profile.Profile {
	gzReader, err := gzip.NewReader(reader)
	require.NoError(t, err, "failed creating gzip reader")

	p, err := profile.Parse(gzReader)
	require.NoError(t, err, "failed parsing profile")

	return p
}

func testMyHeapProfile(t testing.TB) []byte {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	require.NoError(t, pprof.WriteHeapProfile(buf))
	return buf.Bytes()
}

func testStaticCPUProfile(t testing.TB) []byte {
	t.Helper()

	b, err := os.ReadFile("../firedb/testdata/profile")
	require.NoError(t, err)
	return b
}

func BenchmarkIngestProfiles(b *testing.B) {
	cfg := defaultIngesterTestConfig(b)
	logger := log.NewLogfmtLogger(os.Stdout)

	profileStoreCfg := defaultProfileStoreTestConfig(b)
	profileStore, err := profilestore.New(logger, nil, trace.NewNoopTracerProvider(), profileStoreCfg)
	require.NoError(b, err)

	i, err := New(cfg, log.NewNopLogger(), nil, profileStore)
	require.NoError(b, err)

	rawProfile := testStaticCPUProfile(b)
	for n := 0; n < b.N; n++ {
		resp, err := i.Push(context.Background(), connect.NewRequest(&pushv1.PushRequest{
			Series: []*pushv1.RawProfileSeries{
				{
					Labels: []*commonv1.LabelPair{
						{Name: "__name__", Value: "my_own_profile"},
						{Name: "cluster", Value: "us-central1"},
						{Name: "n", Value: fmt.Sprintf("%d", b.N)},
					},
					Samples: []*pushv1.RawSample{
						{
							RawProfile: rawProfile,
						},
					},
				},
			},
		}))
		require.NoError(b, err)
		require.NotNil(b, resp)
	}

	require.NoError(b, profileStore.Table().RotateBlock())

	require.Eventually(b, func() bool {
		files, err := ioutil.ReadDir(profileStoreCfg.DataPath + "/profilestore-v1/fire/stacktraces")
		if err != nil {
			return false
		}

		if len(files) == 0 {
			return false
		}

		if !files[0].IsDir() {
			return false
		}

		stat, err := os.Stat(profileStoreCfg.DataPath + "/profilestore-v1/fire/stacktraces/" + files[0].Name() + "/data.parquet")
		if err != nil {
			return false
		}
		b.Logf("bytes_total=%d bytes_per_profile=%2f", stat.Size(), float64(stat.Size())/float64(b.N))
		return true

	}, time.Second, 50*time.Millisecond)

}
