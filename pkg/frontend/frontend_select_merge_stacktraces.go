package frontend

import (
	"context"

	"github.com/bufbuild/connect-go"

	querierv1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
	"github.com/grafana/phlare/pkg/util/connectgrpc"
)

func (f *Frontend) SelectMergeStacktraces(ctx context.Context, c *connect.Request[querierv1.SelectMergeStacktracesRequest]) (*connect.Response[querierv1.SelectMergeStacktracesResponse], error) {
	return connectgrpc.RoundTripUnary[querierv1.SelectMergeStacktracesRequest, querierv1.SelectMergeStacktracesResponse](ctx, f, c)
}
