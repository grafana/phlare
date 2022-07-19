package fire

import (
	"bytes"
	"context"
	"flag"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/grafana/dskit/kv"
	"github.com/grafana/dskit/kv/consul"
	"github.com/grafana/dskit/modules"
	"github.com/grafana/dskit/ring"
	"github.com/grafana/dskit/services"
	"github.com/grafana/fire/pkg/cfg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	pushv1 "github.com/grafana/fire/pkg/gen/push/v1"
)

func testProfile(t *testing.T) []byte {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	require.NoError(t, pprof.WriteHeapProfile(buf))
	return buf.Bytes()
}

func TestIngestion(t *testing.T) {
	var config Config
	require.NoError(t, cfg.DynamicUnmarshal(&config, []string{"-target", "distributor,ingester"}, flag.CommandLine))

	f, err := New(config)
	require.NoError(t, err)

	initRing := func() (services.Service, error) {
		inmem, closer := consul.NewInMemoryClient(ring.GetCodec(), log.NewNopLogger(), nil)
		t.Cleanup(func() { assert.NoError(t, closer.Close()) })

		cfg := ring.Config{
			KVStore:           kv.Config{Mock: inmem},
			HeartbeatTimeout:  1 * time.Minute,
			ReplicationFactor: 1,
		}

		var err error
		f.ring, err = ring.New(cfg, "ingester", "ring", f.logger, prometheus.WrapRegistererWithPrefix("fire_", f.reg))
		if err != nil {
			return nil, err
		}
		f.Server.HTTP.Path("/ring").Methods("GET", "POST").Handler(f.ring)

		f.Cfg.Ingester.LifecyclerConfig.JoinAfter = time.Millisecond
		f.Cfg.Ingester.LifecyclerConfig.MinReadyDuration = time.Millisecond
		f.Cfg.Ingester.LifecyclerConfig.RingConfig.KVStore.Mock = inmem

		return f.ring, nil
	}

	// ensure we skip the ring, memberlist, distributor setup
	doNothing := func() (services.Service, error) { return nil, nil }
	f.ModuleManager.RegisterModule(MemberlistKV, doNothing, modules.UserInvisibleModule)
	f.ModuleManager.RegisterModule(Ring, initRing, modules.UserInvisibleModule)
	require.NoError(t, f.ModuleManager.AddDependency(Ring, Server))

	// start it
	go func() {
		require.NoError(t, f.Run())
	}()

	// wait for healthy
	require.Eventually(t, func() bool {
		if f.serviceManager == nil {
			return false
		}
		f.serviceManager.AwaitHealthy(context.TODO())
		return true
	}, time.Minute, time.Millisecond)
	t.Log("fire healthy")
	t.Cleanup(func() {
		f.serviceManager.StopAsync()
		f.serviceManager.AwaitStopped(context.TODO())
		t.Log("fire stopped")
	})

	req := connect.NewRequest(&pushv1.PushRequest{
		Series: []*pushv1.RawProfileSeries{
			{
				Labels: []*commonv1.LabelPair{
					{Name: "cluster", Value: "us-central1"},
				},
				Samples: []*pushv1.RawSample{
					{
						RawProfile: testProfile(t),
					},
				},
			},
		},
	})

	var (
		wg     sync.WaitGroup
		ch     = make(chan struct{}, 1)
		worker = func() {
			defer wg.Done()
			for range ch {
				resp, err := f.pusherClient.Push(context.Background(), req)
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		}
	)

	wg.Add(4)
	go worker()
	go worker()
	go worker()
	go worker()

	for _, x := range [100]struct{}{} {
		ch <- x
	}

	close(ch)
	wg.Wait()

}
