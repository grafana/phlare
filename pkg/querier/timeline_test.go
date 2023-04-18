package querier_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	"github.com/grafana/phlare/pkg/querier"
)

const TimelineStepSec = 10

var (
	TestDate = time.Date(2023, time.April, 18, 1, 2, 3, 4, time.UTC)
)

func Test_No_Backfill(t *testing.T) {
	points := &typesv1.Series{
		Points: []*typesv1.Point{
			{Timestamp: TestDate.UnixMilli(), Value: 99},
		},
	}

	timeline := querier.NewTimeline(points, TestDate.UnixMilli(), TestDate.UnixMilli(), TimelineStepSec)

	assert.Equal(t, TestDate.UnixMilli()/1000, timeline.StartTime)
	assert.Equal(t, []uint64{
		99,
	}, timeline.Samples)
}

func Test_Backfill_Data_Start_End(t *testing.T) {
	startTime := TestDate.Add(-1 * time.Minute).UnixMilli()
	endTime := TestDate.Add(1 * time.Minute).UnixMilli()

	points := &typesv1.Series{
		Points: []*typesv1.Point{
			{Timestamp: TestDate.UnixMilli(), Value: 99},
		},
	}

	timeline := querier.NewTimeline(points, startTime, endTime, TimelineStepSec)

	assert.Equal(t, startTime/1000, timeline.StartTime)
	assert.Equal(t, []uint64{
		// 1 point for each 10 seconds
		0, 0, 0, 0, 0,
		99,
		0, 0, 0, 0, 0,
	}, timeline.Samples)
}

func Test_Backfill_Data_Middle(t *testing.T) {
	startTime := TestDate.Add(-1 * time.Minute).UnixMilli()
	endTime := TestDate.Add(1 * time.Minute).UnixMilli()

	points := &typesv1.Series{
		Points: []*typesv1.Point{
			{Timestamp: TestDate.UnixMilli(), Value: 99},
			{Timestamp: TestDate.Add(20 * time.Second).UnixMilli(), Value: 98},
		},
	}

	timeline := querier.NewTimeline(points, startTime, endTime, TimelineStepSec)

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

func Test_Backfill_All(t *testing.T) {
	startTime := TestDate.Add(-1 * time.Minute).UnixMilli()
	endTime := TestDate.Add(1 * time.Minute).UnixMilli()

	points := &typesv1.Series{
		Points: []*typesv1.Point{},
	}

	timeline := querier.NewTimeline(points, startTime, endTime, TimelineStepSec)

	assert.Equal(t, startTime/1000, timeline.StartTime)
	assert.Equal(t, []uint64{
		// 1 point for each 10 seconds
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		// TODO: is this correct?
		0,
	}, timeline.Samples)
}
