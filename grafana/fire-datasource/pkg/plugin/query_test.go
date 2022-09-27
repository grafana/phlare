package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	v1 "github.com/grafana/fire/pkg/gen/common/v1"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/require"

	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	querierv1 "github.com/grafana/fire/pkg/gen/querier/v1"
)

// This is where the tests for the datasource backend live.
func Test_query(t *testing.T) {
	ds := &FireDatasource{
		client: &FakeClient{},
	}

	dataQuery := backend.DataQuery{
		RefID:         "A",
		QueryType:     queryTypeBoth,
		MaxDataPoints: 0,
		Interval:      0,
		TimeRange: backend.TimeRange{
			From: time.UnixMilli(10000),
			To:   time.UnixMilli(20000),
		},
		JSON: []byte(`{"profileTypeId":"foo:bar","labelSelector":"{app=\\\"baz\\\"}"}`),
	}

	t.Run("query both", func(t *testing.T) {
		resp := ds.query(context.Background(), backend.PluginContext{}, dataQuery)
		require.Nil(t, resp.Error)
		require.Equal(t, 2, len(resp.Frames))
		require.Equal(t, "time", resp.Frames[0].Fields[0].Name)
		require.Equal(t, data.NewField("level", nil, []int64{0, 1, 2}), resp.Frames[1].Fields[0])
	})

	t.Run("query profile", func(t *testing.T) {
		dataQuery.QueryType = queryTypeProfile
		resp := ds.query(context.Background(), backend.PluginContext{}, dataQuery)
		require.Nil(t, resp.Error)
		require.Equal(t, 1, len(resp.Frames))
		require.Equal(t, data.NewField("level", nil, []int64{0, 1, 2}), resp.Frames[0].Fields[0])
	})

	t.Run("query metrics", func(t *testing.T) {
		dataQuery.QueryType = queryTypeMetrics
		resp := ds.query(context.Background(), backend.PluginContext{}, dataQuery)
		require.Nil(t, resp.Error)
		require.Equal(t, 1, len(resp.Frames))
		require.Equal(t, "time", resp.Frames[0].Fields[0].Name)
	})
}

// This is where the tests for the datasource backend live.
func Test_profileToDataFrame(t *testing.T) {
	resp := &connect.Response[querierv1.SelectMergeStacktracesResponse]{
		Msg: &querierv1.SelectMergeStacktracesResponse{
			Flamegraph: &querierv1.FlameGraph{
				Names: []string{"func1", "func2", "func3"},
				Levels: []*querierv1.Level{
					{Values: []int64{0, 20, 0, 0}},
					{Values: []int64{0, 10, 0, 1, 0, 5, 0, 2}},
				},
				Total:   987,
				MaxSelf: 123,
			},
		},
	}
	frame := responseToDataFrames(resp, "memory:alloc_objects:count:space:bytes")
	require.Equal(t, 3, len(frame.Fields))
	require.Equal(t, data.NewField("level", nil, []int64{0, 1, 1}), frame.Fields[0])
	require.Equal(t, data.NewField("value", nil, []int64{20, 10, 5}), frame.Fields[1])
	require.Equal(t, data.NewField("label", nil, []string{"func1", "func2", "func3"}), frame.Fields[2])
	require.Equal(t, "memory:alloc_objects:count:space:bytes", frame.Meta.Custom.(CustomMeta).ProfileTypeID)
}

// This is where the tests for the datasource backend live.
func Test_levelsToTree(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		levels := []*querierv1.Level{
			{Values: []int64{0, 100, 0, 0}},
			{Values: []int64{0, 40, 0, 1, 0, 30, 0, 2}},
			{Values: []int64{0, 15, 0, 3}},
		}

		tree := levelsToTree(levels, []string{"root", "func1", "func2", "func1:func3"})
		require.Equal(t, &ProfileTree{
			Start: 0, Value: 100, Level: 0, Name: "root", Nodes: []*ProfileTree{
				{
					Start: 0, Value: 40, Level: 1, Name: "func1", Nodes: []*ProfileTree{
						{Start: 0, Value: 15, Level: 2, Name: "func1:func3"},
					},
				},
				{Start: 40, Value: 30, Level: 1, Name: "func2"},
			},
		}, tree)
	})

	t.Run("medium", func(t *testing.T) {
		levels := []*querierv1.Level{
			{Values: []int64{0, 100, 0, 0}},
			{Values: []int64{0, 40, 0, 1, 0, 30, 0, 2, 0, 30, 0, 3}},
			{Values: []int64{0, 20, 0, 4, 50, 10, 0, 5}},
		}

		tree := levelsToTree(levels, []string{"root", "func1", "func2", "func3", "func1:func4", "func3:func5"})
		require.Equal(t, &ProfileTree{
			Start: 0, Value: 100, Level: 0, Name: "root", Nodes: []*ProfileTree{
				{
					Start: 0, Value: 40, Level: 1, Name: "func1", Nodes: []*ProfileTree{
						{Start: 0, Value: 20, Level: 2, Name: "func1:func4"},
					},
				},
				{Start: 40, Value: 30, Level: 1, Name: "func2"},
				{
					Start: 70, Value: 30, Level: 1, Name: "func3", Nodes: []*ProfileTree{
						{Start: 70, Value: 10, Level: 2, Name: "func3:func5"},
					},
				},
			},
		}, tree)
	})
}

func Test_treeToNestedDataFrame(t *testing.T) {
	tree := &ProfileTree{
		Start: 0, Value: 100, Level: 0, Name: "root", Nodes: []*ProfileTree{
			{
				Start: 10, Value: 40, Level: 1, Name: "func1",
			},
			{Start: 60, Value: 30, Level: 1, Name: "func2", Nodes: []*ProfileTree{
				{Start: 61, Value: 15, Level: 2, Name: "func1:func3"},
			}},
		},
	}

	frame := treeToNestedSetDataFrame(tree, "memory:alloc_objects:count:space:bytes")
	require.Equal(t,
		[]*data.Field{
			data.NewField("level", nil, []int64{0, 1, 1, 2}),
			data.NewField("value", nil, []int64{100, 40, 30, 15}),
			data.NewField("label", nil, []string{"root", "func1", "func2", "func1:func3"}),
		}, frame.Fields)
	require.Equal(t, "memory:alloc_objects:count:space:bytes", frame.Meta.Custom.(CustomMeta).ProfileTypeID)
}

func Test_seriesToDataFrame(t *testing.T) {
	resp := &connect.Response[querierv1.SelectSeriesResponse]{
		Msg: &querierv1.SelectSeriesResponse{
			Series: []*commonv1.Series{
				{Labels: []*commonv1.LabelPair{}, Points: []*commonv1.Point{{Timestamp: int64(1000), Value: 30}, {Timestamp: int64(2000), Value: 10}}},
			},
		},
	}
	frame := seriesToDataFrame(resp, "process_cpu:samples:count:cpu:nanoseconds")
	require.Equal(t, 2, len(frame.Fields))
	require.Equal(t, data.NewField("time", nil, []time.Time{time.UnixMilli(1000), time.UnixMilli(2000)}), frame.Fields[0])
	require.Equal(t, data.NewField("samples", nil, []float64{30, 10}), frame.Fields[1])

	// with a label pair, the value field should name itself with a label pair name and not the profile type
	resp = &connect.Response[querierv1.SelectSeriesResponse]{
		Msg: &querierv1.SelectSeriesResponse{
			Series: []*commonv1.Series{
				{Labels: []*commonv1.LabelPair{{Name: "app", Value: "bar"}}, Points: []*commonv1.Point{{Timestamp: int64(1000), Value: 30}, {Timestamp: int64(2000), Value: 10}}},
			},
		},
	}
	frame = seriesToDataFrame(resp, "process_cpu:samples:count:cpu:nanoseconds")
	require.Equal(t, data.NewField("app", nil, []float64{30, 10}), frame.Fields[1])
}

type FakeClient struct {
	Req *connect.Request[querierv1.SelectMergeStacktracesRequest]
}

func (f FakeClient) ProfileTypes(ctx context.Context, c *connect.Request[querierv1.ProfileTypesRequest]) (*connect.Response[querierv1.ProfileTypesResponse], error) {
	panic("implement me")
}

func (f FakeClient) LabelValues(ctx context.Context, c *connect.Request[querierv1.LabelValuesRequest]) (*connect.Response[querierv1.LabelValuesResponse], error) {
	panic("implement me")
}

func (f FakeClient) LabelNames(context.Context, *connect.Request[querierv1.LabelNamesRequest]) (*connect.Response[querierv1.LabelNamesResponse], error) {
	panic("implement me")
}

func (f FakeClient) Series(ctx context.Context, c *connect.Request[querierv1.SeriesRequest]) (*connect.Response[querierv1.SeriesResponse], error) {
	return &connect.Response[querierv1.SeriesResponse]{
		Msg: &querierv1.SeriesResponse{
			LabelsSet: []*v1.Labels{{
				Labels: []*v1.LabelPair{
					{
						Name:  "__unit__",
						Value: "cpu",
					},
					{
						Name:  "instance",
						Value: "127.0.0.1",
					},
					{
						Name:  "job",
						Value: "default",
					},
				},
			}},
		},
	}, nil
}

func (f FakeClient) SelectMergeStacktraces(ctx context.Context, c *connect.Request[querierv1.SelectMergeStacktracesRequest]) (*connect.Response[querierv1.SelectMergeStacktracesResponse], error) {
	f.Req = c
	return &connect.Response[querierv1.SelectMergeStacktracesResponse]{
		Msg: &querierv1.SelectMergeStacktracesResponse{
			Flamegraph: &querierv1.FlameGraph{
				Names: []string{"foo", "bar", "baz"},
				Levels: []*querierv1.Level{
					{Values: []int64{0, 10, 0, 0}},
					{Values: []int64{0, 9, 0, 1}},
					{Values: []int64{0, 8, 8, 2}},
				},
				Total:   100,
				MaxSelf: 56,
			},
		},
	}, nil
}

func (f FakeClient) SelectSeries(ctx context.Context, req *connect.Request[querierv1.SelectSeriesRequest]) (*connect.Response[querierv1.SelectSeriesResponse], error) {
	return &connect.Response[querierv1.SelectSeriesResponse]{
		Msg: &querierv1.SelectSeriesResponse{
			Series: []*commonv1.Series{
				{
					Labels: []*v1.LabelPair{{Name: "foo", Value: "bar"}},
					Points: []*commonv1.Point{{Timestamp: int64(1000), Value: 30}, {Timestamp: int64(2000), Value: 10}},
				},
			},
		},
	}, nil
}
