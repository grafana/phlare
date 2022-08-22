package firedb

import (
	"context"
	"flag"
	"sync"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/grafana/dskit/services"
	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	profilev1 "github.com/grafana/fire/pkg/gen/google/v1"
	ingestv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
	"github.com/grafana/fire/pkg/iterator"
	firemodel "github.com/grafana/fire/pkg/model"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	DataPath      string
	BlockDuration time.Duration
}

func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&cfg.DataPath, "firedb.data-path", "./data", "Directory used for local storage.")
	f.DurationVar(&cfg.BlockDuration, "firedb.block-duration", 30*time.Minute, "Block duration.")
}

type FireDB struct {
	services.Service

	cfg    *Config
	reg    prometheus.Registerer
	logger log.Logger
	stopCh chan struct{}

	headLock      sync.RWMutex
	head          *Head
	headMetrics   *headMetrics
	headFlushTime time.Time
}

func New(cfg *Config, logger log.Logger, reg prometheus.Registerer) (*FireDB, error) {
	headMetrics := newHeadMetrics(reg)
	f := &FireDB{
		cfg:         cfg,
		reg:         reg,
		logger:      logger,
		stopCh:      make(chan struct{}, 0),
		headMetrics: headMetrics,
	}
	if _, err := f.initHead(); err != nil {
		return nil, err
	}
	f.Service = services.NewBasicService(f.starting, f.running, f.stopping)
	return f, nil
}

func (f *FireDB) loop() {
	for {

		f.headLock.RLock()
		timeToFlush := f.headFlushTime.Sub(time.Now())
		f.headLock.RUnlock()

		select {
		case <-f.stopCh:
			return
		case <-time.After(timeToFlush):
			if err := f.Flush(context.TODO()); err != nil {
				level.Error(f.logger).Log("msg", "flushing head block failed", "err", err)
			}
		}
	}
}

func (f *FireDB) starting(ctx context.Context) error {
	go f.loop()
	return nil
}

func (f *FireDB) running(ctx context.Context) error {
	select {
	// wait until service is asked to stop
	case <-ctx.Done():
		// stop
		close(f.stopCh)
	}
	return nil
}

func (f *FireDB) stopping(_ error) error {
	return f.head.Flush(context.TODO())
}

func (f *FireDB) Head() *Head {
	f.headLock.RLock()
	defer f.headLock.RUnlock()
	return f.head
}

func (f *FireDB) initHead() (oldHead *Head, err error) {
	f.headLock.Lock()
	defer f.headLock.Unlock()
	oldHead = f.head
	f.headFlushTime = time.Now().UTC().Truncate(f.cfg.BlockDuration).Add(f.cfg.BlockDuration)
	f.head, err = NewHead(f.cfg.DataPath, headWithMetrics(f.headMetrics), HeadWithLogger(f.logger))
	if err != nil {
		return oldHead, err
	}
	return oldHead, nil
}

func (f *FireDB) Flush(ctx context.Context) error {
	oldHead, err := f.initHead()
	if err != nil {
		return err
	}

	if oldHead == nil {
		return nil
	}
	return oldHead.Flush(ctx)
}

func (f *FireDB) Ingest(ctx context.Context, p *profilev1.Profile, id uuid.UUID, externalLabels ...*commonv1.LabelPair) error {
	return f.head.Ingest(ctx, p, id, externalLabels...)
}

func (f *FireDB) SelectProfiles(ctx context.Context, req *ingestv1.SelectProfilesRequest) (iterator.Interface[firemodel.Profile], error) {
	return f.head.SelectProfiles(ctx, req)
}

func (f *FireDB) ProfileTypes(ctx context.Context, req *connect.Request[ingestv1.ProfileTypesRequest]) (*connect.Response[ingestv1.ProfileTypesResponse], error) {
	return f.head.ProfileTypes(ctx, req)
}

func (f *FireDB) LabelValues(ctx context.Context, req *connect.Request[ingestv1.LabelValuesRequest]) (*connect.Response[ingestv1.LabelValuesResponse], error) {
	return f.head.LabelValues(ctx, req)
}

func (f *FireDB) Series(ctx context.Context, req *connect.Request[ingestv1.SeriesRequest]) (*connect.Response[ingestv1.SeriesResponse], error) {
	return f.head.Series(ctx, req)
}
