package tripperwares

import (
	"context"

	connect_go "github.com/bufbuild/connect-go"
	v1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
)

type SplitByInterval struct {
	defaultTripperware
}

func (t *SplitByInterval) SelectMergeStacktraces(ctx context.Context, req *connect_go.Request[v1.SelectMergeStacktracesRequest]) (*connect_go.Response[v1.SelectMergeStacktracesResponse], error) {

}
