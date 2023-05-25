package iter

import (
	"math"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	phlaremodel "github.com/grafana/phlare/pkg/model"
)

var (
	aLabels = phlaremodel.LabelsFromStrings("foo", "a")
	bLabels = phlaremodel.LabelsFromStrings("foo", "b")
	cLabels = phlaremodel.LabelsFromStrings("foo", "c")
)

type profile struct {
	labels    phlaremodel.Labels
	timestamp model.Time
}

func (p profile) Labels() phlaremodel.Labels {
	return p.labels
}

func (p profile) Timestamp() model.Time {
	return p.timestamp
}

func TestMergeIterator(t *testing.T) {
	for _, tt := range []struct {
		name        string
		deduplicate bool
		input       [][]profile
		expected    []profile
	}{
		{
			name:        "deduplicate exact",
			deduplicate: true,
			input: [][]profile{
				{
					{labels: aLabels, timestamp: 1},
					{labels: aLabels, timestamp: 2},
					{labels: aLabels, timestamp: 3},
				},
				{
					{labels: aLabels, timestamp: 1},
					{labels: aLabels, timestamp: 2},
					{labels: aLabels, timestamp: 3},
				},
				{
					{labels: aLabels, timestamp: 1},
					{labels: aLabels, timestamp: 2},
					{labels: aLabels, timestamp: 3},
				},
			},
			expected: []profile{
				{labels: aLabels, timestamp: 1},
				{labels: aLabels, timestamp: 2},
				{labels: aLabels, timestamp: 3},
			},
		},
		{
			name: "no deduplicate",
			input: [][]profile{
				{
					{labels: aLabels, timestamp: 1},
					{labels: aLabels, timestamp: 2},
					{labels: aLabels, timestamp: 3},
				},
				{
					{labels: aLabels, timestamp: 1},
					{labels: aLabels, timestamp: 3},
				},
				{
					{labels: aLabels, timestamp: 2},
				},
			},
			expected: []profile{
				{labels: aLabels, timestamp: 1},
				{labels: aLabels, timestamp: 1},
				{labels: aLabels, timestamp: 2},
				{labels: aLabels, timestamp: 2},
				{labels: aLabels, timestamp: 3},
				{labels: aLabels, timestamp: 3},
			},
		},
		{
			name:        "deduplicate and sort",
			deduplicate: true,
			input: [][]profile{
				{
					{labels: aLabels, timestamp: 1},
					{labels: aLabels, timestamp: 2},
					{labels: aLabels, timestamp: 3},
					{labels: aLabels, timestamp: 4},
				},
				{
					{labels: aLabels, timestamp: 1},
					{labels: cLabels, timestamp: 2},
					{labels: aLabels, timestamp: 3},
				},
				{
					{labels: aLabels, timestamp: 2},
					{labels: bLabels, timestamp: 4},
				},
			},
			expected: []profile{
				{labels: aLabels, timestamp: 1},
				{labels: aLabels, timestamp: 2},
				{labels: cLabels, timestamp: 2},
				{labels: aLabels, timestamp: 3},
				{labels: aLabels, timestamp: 4},
				{labels: bLabels, timestamp: 4},
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			iters := make([]Iterator[profile], len(tt.input))
			for i, input := range tt.input {
				iters[i] = NewSliceIterator(input)
			}
			it := NewMergeIterator(
				profile{timestamp: math.MaxInt64},
				tt.deduplicate,
				iters...)
			actual := []profile{}
			for it.Next() {
				actual = append(actual, it.At())
			}
			require.NoError(t, it.Err())
			require.NoError(t, it.Close())
			require.Equal(t, tt.expected, actual)
		})
	}
}

// todo test timedRangeIterator
