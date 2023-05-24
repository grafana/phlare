package querier

import (
	"context"

	"github.com/grafana/dskit/ring"
	ring_client "github.com/grafana/dskit/ring/client"
	"github.com/grafana/mimir/pkg/storegateway"

	"github.com/grafana/phlare/pkg/ingester/clientpool"
)

type StoreGatewayQueryClient interface {
	MergeProfilesStacktraces(context.Context) clientpool.BidiClientMergeProfilesStacktraces
	MergeProfilesLabels(ctx context.Context) clientpool.BidiClientMergeProfilesLabels
	MergeProfilesPprof(ctx context.Context) clientpool.BidiClientMergeProfilesPprof
}

type StoreGatewayLimits interface {
	StoreGatewayTenantShardSize(userID string) int
}

type StoreGatewayQuerier struct {
	ring   ring.ReadRing
	pool   *ring_client.Pool
	limits StoreGatewayLimits
}

func NewStoreGatewayQuerier(pool *ring_client.Pool, ring ring.ReadRing, limits StoreGatewayLimits) *StoreGatewayQuerier {
	return &StoreGatewayQuerier{
		ring:   ring,
		pool:   pool,
		limits: limits,
	}
}

// forAllIngesters runs f, in parallel, for all ingesters
func forAllStoreGateways[T any](ctx context.Context, tenantID string, storegatewayQuerier *StoreGatewayQuerier, f QueryReplicaFn[T, StoreGatewayQueryClient]) ([]ResponseFromReplica[T], error) {
	replicationSet, err := GetShuffleShardingSubring(storegatewayQuerier.ring, tenantID, storegatewayQuerier.limits).GetReplicationSetForOperation(storegateway.BlocksRead)
	if err != nil {
		return nil, err
	}

	return forGivenReplicationSet(ctx, func(addr string) (StoreGatewayQueryClient, error) {
		client, err := storegatewayQuerier.pool.GetClientFor(addr)
		if err != nil {
			return nil, err
		}
		return client.(StoreGatewayQueryClient), nil
	}, replicationSet, f)
}

// GetShuffleShardingSubring returns the subring to be used for a given user. This function
// should be used both by store-gateway and querier in order to guarantee the same logic is used.
func GetShuffleShardingSubring(ring ring.ReadRing, userID string, limits StoreGatewayLimits) ring.ReadRing {
	shardSize := limits.StoreGatewayTenantShardSize(userID)

	// A shard size of 0 means shuffle sharding is disabled for this specific user,
	// so we just return the full ring so that blocks will be sharded across all store-gateways.
	if shardSize <= 0 {
		return ring
	}

	return ring.ShuffleShard(userID, shardSize)
}
