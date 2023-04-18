package querier

import (
	v1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
)

// NewTimeline generates a Pyroscope Timeline
// It assumes:
// * 10 second step
// * Ordered
// * Non-zero
// * startMs is earlier than the first series value
// * endMs is after the last series value
// TODO: sample
func NewTimeline(series *v1.Series, startMs int64, endMs int64) *flamebearer.FlamebearerTimelineV1 {
	points := series.GetPoints()
	durationDeltaInSec := int64(10)

	firstAvailableData := points[0]
	howManyToBackfillStart := (firstAvailableData.Timestamp - startMs) / 1000 / durationDeltaInSec

	// TODO: what if it returns < 0
	// Backfill with 0s for data that's not available
	backFillStart := make([]uint64, int(howManyToBackfillStart))

	lastAvailableData := points[len(points)-1]
	howManyToBackfillEnd := (endMs - lastAvailableData.Timestamp) / 1000 / durationDeltaInSec
	backFillEnd := make([]uint64, int(howManyToBackfillEnd))

	samples := make([]uint64, len(points))
	for i, p := range points {
		samples[i] = uint64(p.Value)
	}

	samples = append(backFillStart, samples...)
	samples = append(samples, backFillEnd...)

	timeline := &flamebearer.FlamebearerTimelineV1{
		// ms to seconds
		StartTime:     startMs / 1000,
		DurationDelta: durationDeltaInSec,
		// Each point corresponds to 10 seconds
		Samples: samples,
	}

	return timeline
}
