package querier

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/grafana/dskit/kv"
	"github.com/grafana/dskit/ring"
	ring_client "github.com/grafana/dskit/ring/client"
	"github.com/grafana/dskit/services"
	"github.com/grafana/mimir/pkg/storegateway"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/grafana/phlare/pkg/storegateway/clientpool"
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

	services.Service
	// Subservices manager.
	subservices        *services.Manager
	subservicesWatcher *services.FailureWatcher
}

func NewStoreGatewayQuerier(
	gatewayCfg storegateway.Config,
	factory ring_client.PoolFactory,
	limits StoreGatewayLimits,
	logger log.Logger,
	reg prometheus.Registerer,
	clientsOptions ...connect.ClientOption,
) (*StoreGatewayQuerier, error) {
	storesRingCfg := gatewayCfg.ShardingRing.ToRingConfig()
	storesRingBackend, err := kv.NewClient(
		storesRingCfg.KVStore,
		ring.GetCodec(),
		kv.RegistererWithKVName(prometheus.WrapRegistererWithPrefix("pyroscope_", reg), "querier-store-gateway"),
		logger,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create store-gateway ring backend")
	}
	storesRing, err := ring.NewWithStoreClientAndStrategy(storesRingCfg, storegateway.RingNameForClient, storegateway.RingKey, storesRingBackend, ring.NewIgnoreUnhealthyInstancesReplicationStrategy(), prometheus.WrapRegistererWithPrefix("pyroscope_", reg), logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create store-gateway ring client")
	}
	// Disable compression for querier -> store-gateway connections
	clientsOptions = append(clientsOptions, connect.WithAcceptCompression("gzip", nil, nil))
	clientsMetrics := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Namespace:   "pyroscope",
		Name:        "storegateway_clients",
		Help:        "The current number of store-gateway clients in the pool.",
		ConstLabels: map[string]string{"client": "querier"},
	})
	pool := clientpool.NewPool(storesRing, factory, clientsMetrics, logger, clientsOptions...)

	s := &StoreGatewayQuerier{
		ring:   storesRing,
		pool:   pool,
		limits: limits,
	}
	s.subservices, err = services.NewManager(storesRing, pool)
	if err != nil {
		return nil, err
	}

	s.Service = services.NewBasicService(s.starting, s.running, s.stopping)

	return s, nil
}

func (s *StoreGatewayQuerier) starting(ctx context.Context) error {
	s.subservicesWatcher.WatchManager(s.subservices)

	if err := services.StartManagerAndAwaitHealthy(ctx, s.subservices); err != nil {
		return errors.Wrap(err, "unable to start store gateway querier set subservices")
	}

	return nil
}

func (s *StoreGatewayQuerier) running(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-s.subservicesWatcher.Chan():
			return errors.Wrap(err, "store gateway querier set subservice failed")
		}
	}
}

func (s *StoreGatewayQuerier) stopping(_ error) error {
	return services.StopManagerAndAwaitStopped(context.Background(), s.subservices)
}

// forAllStoreGateways runs f, in parallel, for all store-gateways that are part of the replication set for the given tenant.
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
