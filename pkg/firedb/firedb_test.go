package firedb

import (
	"context"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	v1 "github.com/grafana/fire/pkg/firedb/schemas/v1"
	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	googlev1 "github.com/grafana/fire/pkg/gen/google/v1"
	ingestv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
	firemodel "github.com/grafana/fire/pkg/model"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

var cpuProfileGenerator = func(tsNano int64, t testing.TB) (*googlev1.Profile, string) {
	p := parseProfile(t, "testdata/profile")
	p.TimeNanos = tsNano
	return p, "process_cpu"
}

func ingestProfiles(b testing.TB, db *FireDB, generator func(tsNano int64, t testing.TB) (*googlev1.Profile, string), from, to int64, step time.Duration, externalLabels ...*commonv1.LabelPair) {
	b.Helper()
	for i := from; i <= to; i += int64(step) {
		p, name := generator(i, b)
		require.NoError(b, db.Head().Ingest(
			context.Background(), p, uuid.New(), append(externalLabels, &commonv1.LabelPair{Name: model.MetricNameLabel, Value: name})...))
	}
}

func TestIteratorOrder(t *testing.T) {
	testDir := t.TempDir()
	db, err := New(&Config{
		DataPath:      testDir,
		BlockDuration: time.Duration(100000) * time.Minute,
	}, log.NewNopLogger(), nil)
	require.NoError(t, err)
	end := time.Unix(0, int64(15*time.Second))
	start := time.Unix(0, 0)
	step := 15 * time.Second
	ingestProfiles(t, db, cpuProfileGenerator, start.UnixNano(), end.UnixNano(), step, &commonv1.LabelPair{Name: "foo", Value: "a"})
	ingestProfiles(t, db, cpuProfileGenerator, start.UnixNano(), end.UnixNano(), step, &commonv1.LabelPair{Name: "foo", Value: "b"})
	require.NoError(t, db.Flush(context.Background()))
	db.runBlockQuerierSync(context.Background())
	require.NoError(t, db.blockQuerier.queriers[0].open(context.Background()))
	actual := []int64{}
	err = db.blockQuerier.queriers[0].forMatchingProfiles(context.Background(),
		[]*labels.Matcher{{Type: labels.MatchEqual, Name: firemodel.LabelNameProfileType, Value: "process_cpu:cpu:nanoseconds:cpu:nanoseconds"}},
		func(lbs firemodel.Labels, _ model.Fingerprint, _ int, profile *v1.Profile) error {
			t.Log(lbs.WithoutPrivateLabels())
			t.Log(time.Unix(0, profile.TimeNanos))
			actual = append(actual, profile.TimeNanos)
			return nil
		})
	require.Equal(t, []int64{0, 0, int64(15 * time.Second), int64(15 * time.Second)}, actual)
	require.NoError(t, err)
}

func BenchmarkDBSelectProfile(b *testing.B) {
	testDir := b.TempDir()
	end := time.Now()
	start := end.Add(-time.Hour)
	step := 15 * time.Second

	db, err := New(&Config{
		DataPath:      testDir,
		BlockDuration: time.Duration(100000) * time.Minute, // we will manually flush
	}, log.NewNopLogger(), nil)
	require.NoError(b, err)

	ingestProfiles(b, db, cpuProfileGenerator, start.UnixNano(), end.UnixNano(), step)

	require.NoError(b, db.Flush(context.Background()))

	db.runBlockQuerierSync(context.Background())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resp, err := db.SelectProfiles(context.Background(), connect.NewRequest(&ingestv1.SelectProfilesRequest{
			LabelSelector: "{}",
			Type:          mustParseProfileSelector(b, "process_cpu:cpu:nanoseconds:cpu:nanoseconds"),
			Start:         int64(model.TimeFromUnixNano(start.UnixNano())),
			End:           int64(model.TimeFromUnixNano(end.UnixNano())),
		}))
		require.NoError(b, err)
		require.True(b, len(resp.Msg.Profiles) != 0)
	}
}
