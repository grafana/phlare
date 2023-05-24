package frontend

import (
	"context"

	"github.com/bufbuild/connect-go"

	querierv1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
	"github.com/grafana/phlare/pkg/util/connectgrpc"
)

func (f *Frontend) SelectSeries(ctx context.Context, c *connect.Request[querierv1.SelectSeriesRequest]) (*connect.Response[querierv1.SelectSeriesResponse], error) {
	return connectgrpc.RoundTripUnary[querierv1.SelectSeriesRequest, querierv1.SelectSeriesResponse](ctx, f, c)
}
