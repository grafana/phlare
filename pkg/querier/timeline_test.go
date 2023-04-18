package querier_test

import (
	"testing"
	"time"

	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	"github.com/grafana/phlare/pkg/querier"
	"github.com/stretchr/testify/assert"
)

func TestNoBackfillData(t *testing.T) {
	d := time.Date(2023, time.April, 18, 1, 2, 3, 4, time.UTC)

	points := &typesv1.Series{
		Points: []*typesv1.Point{
			{Timestamp: d.UnixMilli(), Value: 99},
		},
	}

	timeline := querier.NewTimeline(points, d.UnixMilli(), d.UnixMilli())

	assert.Equal(t, d.UnixMilli()/1000, timeline.StartTime)
	assert.Equal(t, []uint64{
		99,
	}, timeline.Samples)
}

func TestBackfillData(t *testing.T) {
	d := time.Date(2023, time.April, 18, 1, 2, 3, 4, time.UTC)
	startTime := d.Add(-1 * time.Minute).UnixMilli()
	endTime := d.Add(1 * time.Minute).UnixMilli()

	points := &typesv1.Series{
		Points: []*typesv1.Point{
			{Timestamp: d.UnixMilli(), Value: 99},
		},
	}

	timeline := querier.NewTimeline(points, startTime, endTime)

	assert.Equal(t, startTime/1000, timeline.StartTime)
	assert.Equal(t, []uint64{
		// 1 point for each 10 seconds
		0, 0, 0, 0, 0,
		99,
		0, 0, 0, 0, 0,
	}, timeline.Samples)
}

func TestBackfillData2(t *testing.T) {
	d := time.Date(2023, time.April, 18, 1, 2, 3, 4, time.UTC)
	startTime := d.Add(-1 * time.Minute).UnixMilli()
	endTime := d.Add(1 * time.Minute).UnixMilli()

	points := &typesv1.Series{
		Points: []*typesv1.Point{
			{Timestamp: d.UnixMilli(), Value: 99},
			{Timestamp: d.Add(20 * time.Second).UnixMilli(), Value: 98},
		},
	}

	timeline := querier.NewTimeline(points, startTime, endTime)

	assert.Equal(t, startTime/1000, timeline.StartTime)
	assert.Equal(t, []uint64{
		// 1 point for each 10 seconds
		0, 0, 0, 0, 0,
		99,
		0,
		98,
		0, 0, 0, 0,
	}, timeline.Samples)
}
