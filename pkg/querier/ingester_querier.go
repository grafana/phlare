package querier

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/grafana/dskit/ring"
	ring_client "github.com/grafana/dskit/ring/client"

	ingestv1 "github.com/grafana/phlare/api/gen/proto/go/ingester/v1"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	"github.com/grafana/phlare/pkg/ingester/clientpool"
)

type IngesterQueryClient interface {
	LabelValues(context.Context, *connect.Request[typesv1.LabelValuesRequest]) (*connect.Response[typesv1.LabelValuesResponse], error)
	LabelNames(context.Context, *connect.Request[typesv1.LabelNamesRequest]) (*connect.Response[typesv1.LabelNamesResponse], error)
	ProfileTypes(context.Context, *connect.Request[ingestv1.ProfileTypesRequest]) (*connect.Response[ingestv1.ProfileTypesResponse], error)
	Series(ctx context.Context, req *connect.Request[ingestv1.SeriesRequest]) (*connect.Response[ingestv1.SeriesResponse], error)
	MergeProfilesStacktraces(context.Context) clientpool.BidiClientMergeProfilesStacktraces
	MergeProfilesLabels(ctx context.Context) clientpool.BidiClientMergeProfilesLabels
	MergeProfilesPprof(ctx context.Context) clientpool.BidiClientMergeProfilesPprof
}

// IngesterQuerier helps with querying the ingesters.
type IngesterQuerier struct {
	ring ring.ReadRing
	pool *ring_client.Pool
}

func NewIngesterQuerier(pool *ring_client.Pool, ring ring.ReadRing) *IngesterQuerier {
	return &IngesterQuerier{
		ring: ring,
		pool: pool,
	}
}

// forAllIngesters runs f, in parallel, for all ingesters
func forAllIngesters[T any](ctx context.Context, ingesterQuerier *IngesterQuerier, f QueryReplicaFn[T, IngesterQueryClient]) ([]ResponseFromReplica[T], error) {
	replicationSet, err := ingesterQuerier.ring.GetReplicationSetForOperation(ring.Read)
	if err != nil {
		return nil, err
	}
	return forGivenReplicationSet(ctx, func(addr string) (IngesterQueryClient, error) {
		client, err := ingesterQuerier.pool.GetClientFor(addr)
		if err != nil {
			return nil, err
		}
		return client.(IngesterQueryClient), nil
	}, replicationSet, f)
}
