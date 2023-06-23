package phlaredb

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	profilev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	phlaremodel "github.com/grafana/phlare/pkg/model"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/grafana/phlare/pkg/pprof"
	"github.com/grafana/phlare/pkg/pprof/testhelper"
)

func TestComputeDelta(t *testing.T) {
	delta := newDeltaProfiles()
	builder := testhelper.NewProfileBuilder(1).MemoryProfile()
	builder.ForStacktraceString("a", "b", "c").AddSamples(1, 2, 3, 4)
	builder.ForStacktraceString("a", "b", "c", "d").AddSamples(1, 2, 3, 4)

	profiles, labels := newProfileSchema(builder.Profile, "memory")

	samples := delta.computeDelta(profiles[0], labels[0])
	require.Nil(t, samples)
	samples = delta.computeDelta(profiles[1], labels[1])
	require.Nil(t, samples)
	samples = delta.computeDelta(profiles[2], labels[2])
	require.NotNil(t, samples)
	require.Equal(t, 2, len(samples.StacktraceIDs))
	require.Equal(t, int64(3), samples.Values[0])
	require.Equal(t, int64(3), samples.Values[1])
	samples = delta.computeDelta(profiles[3], labels[3])
	require.NotNil(t, samples)
	require.Equal(t, 2, len(samples.StacktraceIDs))
	require.Equal(t, int64(4), samples.Values[0])
	require.Equal(t, int64(4), samples.Values[1])

	profiles, labels = newProfileSchema(builder.Profile, "memory")
	samples = delta.computeDelta(profiles[0], labels[0])
	require.NotNil(t, samples)
	require.Equal(t, 0, len(samples.StacktraceIDs))
	samples = delta.computeDelta(profiles[1], labels[1])
	require.NotNil(t, samples)
	require.Equal(t, 0, len(samples.StacktraceIDs))
	samples = delta.computeDelta(profiles[2], labels[2])
	require.NotNil(t, samples)
	require.Equal(t, 2, len(samples.StacktraceIDs))
	require.Equal(t, int64(3), samples.Values[0])
	require.Equal(t, int64(3), samples.Values[1])
	samples = delta.computeDelta(profiles[3], labels[3])
	require.NotNil(t, samples)
	require.Equal(t, 2, len(samples.StacktraceIDs))
	require.Equal(t, int64(4), samples.Values[0])
	require.Equal(t, int64(4), samples.Values[1])
}

func newProfileSchema(p *profilev1.Profile, name string) ([]schemav1.InMemoryProfile, []phlaremodel.Labels) {
	var (
		labels, seriesRefs = labelsForProfile(p, &typesv1.LabelPair{Name: model.MetricNameLabel, Value: name})
		ps                 = make([]schemav1.InMemoryProfile, len(labels))
	)
	for idxType := range labels {
		ps[idxType] = schemav1.InMemoryProfile{
			ID:                uuid.New(),
			TimeNanos:         p.TimeNanos,
			Comments:          p.Comment,
			DurationNanos:     p.DurationNanos,
			DropFrames:        p.DropFrames,
			KeepFrames:        p.KeepFrames,
			Period:            p.Period,
			DefaultSampleType: p.DefaultSampleType,
		}
		hashes := pprof.StacktracesHasher{}.Hashes(p.Sample)
		ps[idxType].Samples = schemav1.Samples{
			StacktraceIDs: make([]uint32, len(p.Sample)),
			Values:        make([]uint64, len(p.Sample)),
		}
		for i, s := range p.Sample {
			ps[idxType].Samples.Values[i] = uint64(s.Value[idxType])
			ps[idxType].Samples.StacktraceIDs[i] = uint32(hashes[i])

		}
		ps[idxType].SeriesFingerprint = seriesRefs[idxType]
	}
	return ps, labels
}

func TestDeltaSample(t *testing.T) {
	new := schemav1.Samples{
		StacktraceIDs: []uint32{2, 3},
		Values:        []uint64{1, 1},
	}
	highest := map[uint32]uint64{}
	_ = deltaSamples(highest, new)
	require.Equal(t, 2, len(highest))
	require.Equal(t, []*schemav1.Sample{
		{StacktraceID: 2, Value: 1},
		{StacktraceID: 3, Value: 1},
	}, highest)
	require.Equal(t, highest, new)

	t.Run("same stacktraces, matching counter samples, matching gauge samples", func(t *testing.T) {
		new = schemav1.Samples{
			StacktraceIDs: []uint32{2, 3},
			Values:        []uint64{1, 1},
		}
		_ = deltaSamples(highest, new)
		require.Equal(t, 2, len(highest))
		require.Equal(t, []*schemav1.Sample{
			{StacktraceID: 2, Value: 1},
			{StacktraceID: 3, Value: 1},
		}, highest)
		require.Equal(t, []*schemav1.Sample{
			{StacktraceID: 2, Value: 0},
			{StacktraceID: 3, Value: 0},
		}, new)
	})

	t.Run("same stacktraces, matching counter samples, empty gauge samples", func(t *testing.T) {
		new = schemav1.Samples{
			StacktraceIDs: []uint32{2, 3},
			Values:        []uint64{1, 1},
		}
		_ = deltaSamples(highest, new)
		require.Equal(t, 2, len(highest))
		require.Equal(t, []*schemav1.Sample{
			{StacktraceID: 2, Value: 1},
			{StacktraceID: 3, Value: 1},
		}, highest)
		require.Equal(t, []*schemav1.Sample{
			{StacktraceID: 2, Value: 0},
			{StacktraceID: 3, Value: 0},
		}, new)
	})

	t.Run("new stacktrace, and increase counter in existing stacktrace", func(t *testing.T) {
		new = schemav1.Samples{
			StacktraceIDs: []uint32{3, 5},
			Values:        []uint64{6, 1},
		}
		_ = deltaSamples(highest, new)
		require.Equal(t, []*schemav1.Sample{
			{StacktraceID: 2, Value: 1},
			{StacktraceID: 3, Value: 6},
			{StacktraceID: 5, Value: 1},
		}, highest)
	})

	t.Run("same stacktraces, counter samples resetting", func(t *testing.T) {
		new = schemav1.Samples{
			StacktraceIDs: []uint32{3, 5},
			Values:        []uint64{0, 1},
		}
		reset := deltaSamples(highest, new)
		require.True(t, reset)
		require.Equal(t, []*schemav1.Sample{
			{StacktraceID: 3, Value: 1},
			{StacktraceID: 5, Value: 0},
		}, new)
	})

	t.Run("two new stacktraces, raise counters of existing stacktrace", func(t *testing.T) {
		new = schemav1.Samples{
			StacktraceIDs: []uint32{0, 1, 7},
			Values:        []uint64{10, 2, 1},
		}

		_ = deltaSamples(highest, new)
		// sort.Slice(highest, func(i, j int) bool {
		// 	return highest[i].StacktraceID < highest[j].StacktraceID
		// })
		require.Equal(t, []*schemav1.Sample{
			{StacktraceID: 0, Value: 10},
			{StacktraceID: 1, Value: 2},
			{StacktraceID: 2, Value: 1},
			{StacktraceID: 3, Value: 6},
			{StacktraceID: 5, Value: 1},
			{StacktraceID: 7, Value: 1},
		}, highest)
		require.Equal(t, []*schemav1.Sample{
			{StacktraceID: 0, Value: 10},
			{StacktraceID: 1, Value: 2},
			{StacktraceID: 7, Value: 1},
		}, new)
	})
}
