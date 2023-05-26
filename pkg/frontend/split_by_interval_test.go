package frontend

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/grafana/phlare/pkg/iter"
)

func Test_TimeIntervalIterator(t *testing.T) {
	type testCase struct {
		description string
		inputRange  TimeInterval
		expected    []TimeInterval
	}

	testCases := []testCase{
		{
			description: "misaligned time range",
			inputRange:  TimeInterval{time.Unix(0, 1), time.Unix(0, 3602)},
			expected: []TimeInterval{
				{time.Unix(0, 1), time.Unix(0, 899)},
				{time.Unix(0, 900), time.Unix(0, 1799)},
				{time.Unix(0, 1800), time.Unix(0, 2699)},
				{time.Unix(0, 2700), time.Unix(0, 3599)},
				{time.Unix(0, 3600), time.Unix(0, 3602)},
			},
		},
		{
			description: "round range",
			inputRange:  TimeInterval{time.Unix(0, 0), time.Unix(0, 3600)},
			expected: []TimeInterval{
				{time.Unix(0, 0), time.Unix(0, 899)},
				{time.Unix(0, 900), time.Unix(0, 1799)},
				{time.Unix(0, 1800), time.Unix(0, 2699)},
				{time.Unix(0, 2700), time.Unix(0, 3600)},
			},
		},
		{
			description: "exact range",
			inputRange:  TimeInterval{time.Unix(0, 900), time.Unix(0, 1800)},
			expected:    []TimeInterval{{time.Unix(0, 900), time.Unix(0, 1800)}},
		},
		{
			description: "zero range",
		},
		{
			description: "range less than interval",
			inputRange:  TimeInterval{time.Unix(0, 1), time.Unix(0, 501)},
			expected: []TimeInterval{
				{time.Unix(0, 1), time.Unix(0, 501)},
			},
		},
		{
			description: "range less than interval",
			inputRange:  TimeInterval{time.Unix(0, 1), time.Unix(0, 501)},
			expected: []TimeInterval{
				{time.Unix(0, 1), time.Unix(0, 501)},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			actual, err := iter.Slice[TimeInterval](NewTimeIntervalIterator(tc.inputRange.Start, tc.inputRange.End, 900))
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_TimeIntervalIterator_ZeroInterval(t *testing.T) {
	actual, err := iter.Slice[TimeInterval](NewTimeIntervalIterator(
		time.UnixMilli(51),
		time.UnixMilli(211),
		0))

	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, int64(51), actual[0].Start.UnixMilli())
	require.Equal(t, int64(211), actual[0].End.UnixMilli())
}

func Test_TimeIntervalIterator_Milli(t *testing.T) {
	actual, err := iter.Slice[TimeInterval](NewTimeIntervalIterator(
		time.UnixMilli(51),
		time.UnixMilli(211),
		50*time.Millisecond))

	require.NoError(t, err)
	require.Len(t, actual, 4)
	require.Equal(t, int64(51), actual[0].Start.UnixMilli())
	require.Equal(t, int64(99), actual[0].End.UnixMilli())
	require.Equal(t, int64(100), actual[1].Start.UnixMilli())
	require.Equal(t, int64(149), actual[1].End.UnixMilli())
	require.Equal(t, int64(150), actual[2].Start.UnixMilli())
	require.Equal(t, int64(199), actual[2].End.UnixMilli())
	require.Equal(t, int64(200), actual[3].Start.UnixMilli())
	require.Equal(t, int64(211), actual[3].End.UnixMilli())
}
