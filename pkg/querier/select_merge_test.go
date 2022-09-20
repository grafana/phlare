package querier

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	ingestv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
	"github.com/grafana/fire/pkg/ingester/clientpool"
	"github.com/grafana/fire/pkg/iter"
	firemodel "github.com/grafana/fire/pkg/model"
	"github.com/grafana/fire/pkg/testhelper"
)

var foobarlabels = []*commonv1.LabelPair{{Name: "foo", Value: "bar"}}

func TestSelectMergeStacktraces(t *testing.T) {
	resp1 := newFakeBidiClientStacktraces([]*ingestv1.ProfileSets{
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 1},
				{LabelIndex: 0, Timestamp: 2},
				{LabelIndex: 0, Timestamp: 4},
			},
		},
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 5},
				{LabelIndex: 0, Timestamp: 6},
			},
		},
	})
	resp2 := newFakeBidiClientStacktraces([]*ingestv1.ProfileSets{
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 2},
				{LabelIndex: 0, Timestamp: 3},
				{LabelIndex: 0, Timestamp: 4},
			},
		},
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 5},
				{LabelIndex: 0, Timestamp: 6},
			},
		},
	})
	resp3 := newFakeBidiClientStacktraces([]*ingestv1.ProfileSets{
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 3},
				{LabelIndex: 0, Timestamp: 5},
			},
		},
	})
	res, err := selectMergeStacktraces(context.Background(), []responseFromIngesters[clientpool.BidiClientMergeProfilesStacktraces]{
		{
			response: resp1,
		},
		{
			response: resp2,
		},
		{
			response: resp3,
		},
	})
	require.NoError(t, err)
	require.Len(t, res, 1)
	all := []testProfile{}
	all = append(all, resp1.kept...)
	all = append(all, resp2.kept...)
	all = append(all, resp3.kept...)
	sort.Slice(all, func(i, j int) bool { return all[i].Ts < all[j].Ts })
	testhelper.EqualProto(t, all, []testProfile{
		{Ts: 1, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 2, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 3, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 4, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 5, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 6, Labels: &commonv1.Labels{Labels: foobarlabels}},
	})
	res, err = selectMergeStacktraces(context.Background(), []responseFromIngesters[clientpool.BidiClientMergeProfilesStacktraces]{
		{
			response: newFakeBidiClientStacktraces([]*ingestv1.ProfileSets{
				{
					LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
					Profiles: []*ingestv1.SeriesProfile{
						{LabelIndex: 0, Timestamp: 1},
						{LabelIndex: 0, Timestamp: 2},
						{LabelIndex: 0, Timestamp: 4},
					},
				},
				{
					LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
					Profiles: []*ingestv1.SeriesProfile{
						{LabelIndex: 0, Timestamp: 5},
						{LabelIndex: 0, Timestamp: 6},
					},
				},
			}),
		},
	})
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestSelectMergeByLabels(t *testing.T) {
	resp1 := newFakeBidiClientSeries([]*ingestv1.ProfileSets{
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 1},
				{LabelIndex: 0, Timestamp: 2},
				{LabelIndex: 0, Timestamp: 4},
			},
		},
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 5},
				{LabelIndex: 0, Timestamp: 6},
			},
		},
	}, &commonv1.Series{
		Labels: []*commonv1.LabelPair{{Name: "foo", Value: "bar"}},
		Points: []*commonv1.Point{{T: 1, V: 1.0}, {T: 2, V: 2.0}},
	})
	resp2 := newFakeBidiClientSeries([]*ingestv1.ProfileSets{
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 2},
				{LabelIndex: 0, Timestamp: 3},
				{LabelIndex: 0, Timestamp: 4},
			},
		},
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 5},
				{LabelIndex: 0, Timestamp: 6},
			},
		},
	}, &commonv1.Series{
		Labels: foobarlabels,
		Points: []*commonv1.Point{{T: 3, V: 3.0}, {T: 4, V: 4.0}},
	})
	resp3 := newFakeBidiClientSeries([]*ingestv1.ProfileSets{
		{
			LabelsSets: []*commonv1.Labels{{Labels: foobarlabels}},
			Profiles: []*ingestv1.SeriesProfile{
				{LabelIndex: 0, Timestamp: 3},
				{LabelIndex: 0, Timestamp: 5},
			},
		},
	}, &commonv1.Series{
		Labels: foobarlabels,
		Points: []*commonv1.Point{{T: 5, V: 5.0}, {T: 6, V: 6.0}},
	})

	res, err := selectMergeSeries(context.Background(), []responseFromIngesters[clientpool.BidiClientMergeProfilesLabels]{
		{
			response: resp1,
		},
		{
			response: resp2,
		},
		{
			response: resp3,
		},
	})
	require.NoError(t, err)
	// ensure we have correctly selected the right profiles
	all := []testProfile{}
	all = append(all, resp1.kept...)
	all = append(all, resp2.kept...)
	all = append(all, resp3.kept...)
	sort.Slice(all, func(i, j int) bool { return all[i].Ts < all[j].Ts })
	testhelper.EqualProto(t, all, []testProfile{
		{Ts: 1, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 2, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 3, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 4, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 5, Labels: &commonv1.Labels{Labels: foobarlabels}},
		{Ts: 6, Labels: &commonv1.Labels{Labels: foobarlabels}},
	})
	values, err := iter.Slice(res)
	require.NoError(t, err)
	require.Equal(t, []ProfileValue{
		{ts: 1, Value: 1.0, lbs: foobarlabels, LabelsHash: firemodel.Labels(foobarlabels).Hash()},
		{ts: 2, Value: 2.0, lbs: foobarlabels, LabelsHash: firemodel.Labels(foobarlabels).Hash()},
		{ts: 3, Value: 3.0, lbs: foobarlabels, LabelsHash: firemodel.Labels(foobarlabels).Hash()},
		{ts: 4, Value: 4.0, lbs: foobarlabels, LabelsHash: firemodel.Labels(foobarlabels).Hash()},
		{ts: 5, Value: 5.0, lbs: foobarlabels, LabelsHash: firemodel.Labels(foobarlabels).Hash()},
		{ts: 6, Value: 6.0, lbs: foobarlabels, LabelsHash: firemodel.Labels(foobarlabels).Hash()},
	}, values)
}