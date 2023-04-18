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
func NewTimeline(series *v1.Series, startMs int64, endMs int64) *flamebearer.FlamebearerTimelineV1 {
	points := series.GetPoints()
	durationDeltaInSec := int64(10)

	firstAvailableData := points[0]
	lastAvailableData := points[len(points)-1]

	// TODO: what if it returns < 0
	// Backfill with 0s for data that's not available
	backFillStart := backfill(startMs, firstAvailableData.Timestamp, durationDeltaInSec)
	backFillEnd := backfill(lastAvailableData.Timestamp, endMs, durationDeltaInSec)

	samples := make([]uint64, len(points))

	i := 0
	prev := points[0]
	for _, p := range points {
		backfillNum := needsToBackfill(prev.Timestamp, p.Timestamp, durationDeltaInSec)

		if backfillNum > 0 {
			// backfill + newValue
			bf := append(backfill(prev.Timestamp, p.Timestamp, durationDeltaInSec), uint64(p.Value))

			// break the slice
			first := samples[:i]
			second := samples[i:]

			// add new backfilled items
			first = append(first, bf...)

			// concatenate the three slices to form the new slice
			samples = append(first, second...)
			prev = p
			i = i + int(backfillNum)
		} else {
			samples[i] = uint64(p.Value)
			prev = p
			i = i + 1
		}
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

func needsToBackfill(startMs int64, endMs int64, stepSec int64) int64 {
	startSec := startMs / 1000
	endSec := endMs / 1000
	size := ((endSec - startSec) - stepSec) / stepSec
	return size
}

func backfill(startMs int64, endMs int64, stepSec int64) []uint64 {
	startSec := startMs / 1000
	endSec := endMs / 1000

	size := ((endSec - startSec) - stepSec) / stepSec
	if size <= 0 {
		size = 0
	}
	return make([]uint64, size)
}
