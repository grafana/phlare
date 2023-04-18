package querier

import (
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"

	v1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
)

// NewTimeline generates a Pyroscope Timeline
// It assumes:
// * Ordered
// * startMs is earlier than the first series value
// * endMs is after the last series value
func NewTimeline(series *v1.Series, startMs int64, endMs int64, durationDeltaSec int64) *flamebearer.FlamebearerTimelineV1 {
	// ms to seconds
	startSec := startMs / 1000

	points := series.GetPoints()

	if len(points) < 1 {
		return &flamebearer.FlamebearerTimelineV1{
			StartTime:     startSec,
			DurationDelta: durationDeltaSec,
			Samples:       backfill(startMs, endMs, durationDeltaSec),
		}
	}

	firstAvailableData := points[0]
	lastAvailableData := points[len(points)-1]

	// Backfill with 0s for data that's not available
	backFillStart := backfill(startMs, firstAvailableData.Timestamp, durationDeltaSec)
	backFillEnd := backfill(lastAvailableData.Timestamp, endMs, durationDeltaSec)

	samples := make([]uint64, len(points))

	i := 0
	prev := points[0]
	for _, p := range points {
		backfillNum := sizeToBackfill(prev.Timestamp, p.Timestamp, durationDeltaSec)

		if backfillNum > 0 {
			// backfill + newValue
			bf := append(backfill(prev.Timestamp, p.Timestamp, durationDeltaSec), uint64(p.Value))

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
		StartTime:     startSec,
		DurationDelta: durationDeltaSec,
		Samples:       samples,
	}

	return timeline
}

// sizeToBackfill indicates how many items are needed to backfill
// if none are needed, a negative value is returned
func sizeToBackfill(startMs int64, endMs int64, stepSec int64) int64 {
	startSec := startMs / 1000
	endSec := endMs / 1000
	size := ((endSec - startSec) - stepSec) / stepSec
	return size
}

func backfill(startMs int64, endMs int64, stepSec int64) []uint64 {
	size := sizeToBackfill(startMs, endMs, stepSec)
	if size <= 0 {
		size = 0
	}
	return make([]uint64, size)
}
