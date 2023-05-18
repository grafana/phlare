package tripperwares

import (
	"context"

	connect_go "github.com/bufbuild/connect-go"
	v12 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	v1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
	v11 "github.com/grafana/phlare/api/gen/proto/go/types/v1"

	"github.com/grafana/phlare/api/gen/proto/go/querier/v1/querierv1connect"
)

type defaultTripperware struct {
	next querierv1connect.QuerierServiceHandler
}

func (dt *defaultTripperware) ProfileTypes(ctx context.Context, req *connect_go.Request[v1.ProfileTypesRequest]) (*connect_go.Response[v1.ProfileTypesResponse], error) {
	return dt.next.ProfileTypes(ctx, req)
}

func (dt *defaultTripperware) LabelValues(ctx context.Context, req *connect_go.Request[v11.LabelValuesRequest]) (*connect_go.Response[v11.LabelValuesResponse], error) {
	return dt.next.LabelValues(ctx, req)
}

func (dt *defaultTripperware) LabelNames(ctx context.Context, req *connect_go.Request[v11.LabelNamesRequest]) (*connect_go.Response[v11.LabelNamesResponse], error) {
	return dt.next.LabelNames(ctx, req)
}

func (dt *defaultTripperware) Series(ctx context.Context, req *connect_go.Request[v1.SeriesRequest]) (*connect_go.Response[v1.SeriesResponse], error) {
	return dt.next.Series(ctx, req)
}

func (dt *defaultTripperware) SelectMergeStacktraces(ctx context.Context, req *connect_go.Request[v1.SelectMergeStacktracesRequest]) (*connect_go.Response[v1.SelectMergeStacktracesResponse], error) {
	return dt.next.SelectMergeStacktraces(ctx, req)
}

func (dt *defaultTripperware) SelectMergeProfile(ctx context.Context, req *connect_go.Request[v1.SelectMergeProfileRequest]) (*connect_go.Response[v12.Profile], error) {
	return dt.next.SelectMergeProfile(ctx, req)
}

func (dt *defaultTripperware) SelectSeries(ctx context.Context, req *connect_go.Request[v1.SelectSeriesRequest]) (*connect_go.Response[v1.SelectSeriesResponse], error) {
	return dt.next.SelectSeries(ctx, req)
}

func (dt *defaultTripperware) Diff(ctx context.Context, req *connect_go.Request[v1.DiffRequest]) (*connect_go.Response[v1.DiffResponse], error) {
	return dt.next.Diff(ctx, req)
}
