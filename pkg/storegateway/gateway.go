package storegateway

import (
	"github.com/go-kit/log"
	"github.com/grafana/dskit/kv"
	"github.com/grafana/dskit/ring"
	"github.com/grafana/dskit/services"
	"github.com/grafana/mimir/pkg/storegateway"
	"github.com/grafana/mimir/pkg/util/activitytracker"
	"github.com/grafana/phlare/pkg/validation"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type StoreGateway struct {
	services.Service
	logger log.Logger

	gatewayCfg storegateway.Config

	// Ring used for sharding blocks.
	ringLifecycler *ring.BasicLifecycler
	ring           *ring.Ring

	// Subservices manager (ring, lifecycler)
	subservices        *services.Manager
	subservicesWatcher *services.FailureWatcher
}



func NewStoreGateway(gatewayCfg storegateway.Config, storageCfg mimir_tsdb.BlocksStorageConfig, limits *validation.Overrides, logger log.Logger, reg prometheus.Registerer, tracker *activitytracker.ActivityTracker) (*StoreGateway, error) {
	var ringStore kv.Client

	bucketClient, err := createBucketClient(storageCfg, logger, reg)
	if err != nil {
		return nil, err
	}

	ringStore, err = kv.NewClient(
		gatewayCfg.ShardingRing.KVStore,
		ring.GetCodec(),
		kv.RegistererWithKVName(prometheus.WrapRegistererWithPrefix("cortex_", reg), "store-gateway"),
		logger,
	)
	if err != nil {
		return nil, errors.Wrap(err, "create KV store client")
	}

	return newStoreGateway(gatewayCfg, storageCfg, bucketClient, ringStore, limits, logger, reg, tracker)
}
