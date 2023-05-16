package storegateway

import (
	"github.com/go-kit/log"
	"github.com/grafana/dskit/ring"
	"github.com/grafana/dskit/services"
	"github.com/grafana/mimir/pkg/storegateway"
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
