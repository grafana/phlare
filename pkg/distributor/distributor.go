package distributor

import (
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"hash/fnv"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/grafana/dskit/ring"
	ring_client "github.com/grafana/dskit/ring/client"
	"github.com/grafana/dskit/services"
	"github.com/opentracing/opentracing-go"
	parcastorev1 "github.com/parca-dev/parca/gen/proto/go/parca/profilestore/v1alpha1"
	"github.com/parca-dev/parca/pkg/scrape"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/weaveworks/common/user"
	"go.uber.org/atomic"

	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	pushv1 "github.com/grafana/fire/pkg/gen/push/v1"
	"github.com/grafana/fire/pkg/ingester/clientpool"
	firemodel "github.com/grafana/fire/pkg/model"
	"github.com/grafana/fire/pkg/pprof"
)

type PushClient interface {
	Push(context.Context, *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error)
}

// todo: move to non global metrics.
var clients = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "fire",
	Name:      "distributor_ingester_clients",
	Help:      "The current number of ingester clients.",
})

// Config for a Distributor.
type Config struct {
	PushTimeout time.Duration
	PoolConfig  clientpool.PoolConfig `yaml:"pool_config,omitempty"`
}

// RegisterFlags registers distributor-related flags.
func (cfg *Config) RegisterFlags(fs *flag.FlagSet) {
	cfg.PoolConfig.RegisterFlagsWithPrefix("distributor", fs)
	fs.DurationVar(&cfg.PushTimeout, "distributor.push.timeout", 5*time.Second, "Timeout when pushing data to ingester.")
}

// Distributor coordinates replicates and distribution of log streams.
type Distributor struct {
	services.Service
	logger log.Logger

	cfg           Config
	ingestersRing ring.ReadRing
	pool          *ring_client.Pool

	subservices        *services.Manager
	subservicesWatcher *services.FailureWatcher

	metrics *metrics
}

func New(cfg Config, ingestersRing ring.ReadRing, factory ring_client.PoolFactory, reg prometheus.Registerer, logger log.Logger) (*Distributor, error) {
	d := &Distributor{
		cfg:           cfg,
		logger:        logger,
		ingestersRing: ingestersRing,
		pool:          clientpool.NewPool(cfg.PoolConfig, ingestersRing, factory, clients, logger),
		metrics:       newMetrics(reg),
	}
	var err error
	d.subservices, err = services.NewManager(d.pool)
	if err != nil {
		return nil, errors.Wrap(err, "services manager")
	}
	d.subservicesWatcher = services.NewFailureWatcher()
	d.subservicesWatcher.WatchManager(d.subservices)
	d.Service = services.NewBasicService(d.starting, d.running, d.stopping)
	return d, nil
}

func (d *Distributor) starting(ctx context.Context) error {
	return services.StartManagerAndAwaitHealthy(ctx, d.subservices)
}

func (d *Distributor) running(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-d.subservicesWatcher.Chan():
		return errors.Wrap(err, "distributor subservice failed")
	}
}

func (d *Distributor) stopping(_ error) error {
	return services.StopManagerAndAwaitStopped(context.Background(), d.subservices)
}

func (d *Distributor) Push(ctx context.Context, req *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error) {
	var (
		keys     = make([]uint32, 0, len(req.Msg.Series))
		profiles = make([]*profileTracker, 0, len(req.Msg.Series))

		// todo pool readers/writer
		gzipReader *gzip.Reader
		gzipWriter *gzip.Writer
		err        error
		br         = bytes.NewReader(nil)
	)

	for _, series := range req.Msg.Series {
		// todo propagate tenantID.
		keys = append(keys, TokenFor("", labelsString(series.Labels)))
		profName := firemodel.Labels(series.Labels).Get(scrape.ProfileName)
		for _, raw := range series.Samples {
			d.metrics.receivedCompressedBytes.WithLabelValues(profName).Observe(float64(len(raw.RawProfile)))
			br.Reset(raw.RawProfile)
			if gzipReader == nil {
				gzipReader, err = gzip.NewReader(br)
				if err != nil {
					return nil, errors.Wrap(err, "gzip reader")
				}
			} else {
				if err := gzipReader.Reset(br); err != nil {
					return nil, errors.Wrap(err, "gzip reset")
				}
			}
			data, err := ioutil.ReadAll(gzipReader)
			if err != nil {
				return nil, errors.Wrap(err, "gzip read all")
			}
			d.metrics.receivedDecompressedBytes.WithLabelValues(profName).Observe(float64(len(data)))
			p, err := pprof.OpenRaw(data)
			if err != nil {
				return nil, err
			}
			d.metrics.receivedSamples.WithLabelValues(profName).Observe(float64(len(p.Sample)))

			p.Normalize()

			level.Warn(d.logger).Log("msg", "received sample", "labels", firemodel.LabelPairsString(series.Labels), "type", p.StringTable[p.SampleType[0].Type])

			// reuse the data buffer if possible
			size := p.SizeVT()
			if cap(data) < size {
				data = make([]byte, size)
			}
			n, err := p.MarshalToVT(data)
			if err != nil {
				return nil, err
			}
			p.ReturnToVTPool()
			data = data[:n]

			// zip the data back into the buffer
			bw := bytes.NewBuffer(raw.RawProfile[:0])
			if gzipWriter == nil {
				gzipWriter = gzip.NewWriter(bw)
			} else {
				gzipWriter.Reset(bw)
			}
			if _, err := gzipWriter.Write(data); err != nil {
				return nil, errors.Wrap(err, "gzip write")
			}
			if err := gzipWriter.Close(); err != nil {
				return nil, errors.Wrap(err, "gzip close")
			}
			raw.RawProfile = bw.Bytes()
			// generate a unique profile ID before pushing.
			raw.ID = uuid.NewString()
		}
		profiles = append(profiles, &profileTracker{profile: series})
	}

	const maxExpectedReplicationSet = 5 // typical replication factor 3 plus one for inactive plus one for luck
	var descs [maxExpectedReplicationSet]ring.InstanceDesc

	samplesByIngester := map[string][]*profileTracker{}
	ingesterDescs := map[string]ring.InstanceDesc{}
	for i, key := range keys {
		replicationSet, err := d.ingestersRing.Get(key, ring.Write, descs[:0], nil, nil)
		if err != nil {
			return nil, err
		}
		profiles[i].minSuccess = len(replicationSet.Instances) - replicationSet.MaxErrors
		profiles[i].maxFailures = replicationSet.MaxErrors
		for _, ingester := range replicationSet.Instances {
			samplesByIngester[ingester.Addr] = append(samplesByIngester[ingester.Addr], profiles[i])
			ingesterDescs[ingester.Addr] = ingester
		}
	}
	tracker := pushTracker{
		done: make(chan struct{}, 1), // buffer avoids blocking if caller terminates - sendProfiles() only sends once on each
		err:  make(chan error, 1),
	}
	tracker.samplesPending.Store(int32(len(profiles)))
	for ingester, samples := range samplesByIngester {
		go func(ingester ring.InstanceDesc, samples []*profileTracker) {
			// Use a background context to make sure all ingesters get samples even if we return early
			localCtx, cancel := context.WithTimeout(context.Background(), d.cfg.PushTimeout)
			defer cancel()
			localCtx = user.InjectOrgID(localCtx, "")
			if sp := opentracing.SpanFromContext(ctx); sp != nil {
				localCtx = opentracing.ContextWithSpan(localCtx, sp)
			}
			d.sendProfiles(localCtx, ingester, samples, &tracker)
		}(ingesterDescs[ingester], samples)
	}
	select {
	case err := <-tracker.err:
		return nil, err
	case <-tracker.done:
		return connect.NewResponse(&pushv1.PushResponse{}), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (d *Distributor) sendProfiles(ctx context.Context, ingester ring.InstanceDesc, profileTrackers []*profileTracker, pushTracker *pushTracker) {
	err := d.sendProfilesErr(ctx, ingester, profileTrackers)
	// If we succeed, decrement each sample's pending count by one.  If we reach
	// the required number of successful puts on this sample, then decrement the
	// number of pending samples by one.  If we successfully push all samples to
	// min success ingesters, wake up the waiting rpc so it can return early.
	// Similarly, track the number of errors, and if it exceeds maxFailures
	// shortcut the waiting rpc.
	//
	// The use of atomic increments here guarantees only a single sendSamples
	// goroutine will write to either channel.
	for i := range profileTrackers {
		if err != nil {
			if profileTrackers[i].failed.Inc() <= int32(profileTrackers[i].maxFailures) {
				continue
			}
			if pushTracker.samplesFailed.Inc() == 1 {
				pushTracker.err <- err
			}
		} else {
			if profileTrackers[i].succeeded.Inc() != int32(profileTrackers[i].minSuccess) {
				continue
			}
			if pushTracker.samplesPending.Dec() == 0 {
				pushTracker.done <- struct{}{}
			}
		}
	}
}

func (d *Distributor) sendProfilesErr(ctx context.Context, ingester ring.InstanceDesc, profileTrackers []*profileTracker) error {
	c, err := d.pool.GetClientFor(ingester.Addr)
	if err != nil {
		return err
	}

	req := connect.NewRequest(&pushv1.PushRequest{
		Series: make([]*pushv1.RawProfileSeries, 0, len(profileTrackers)),
	})

	for _, p := range profileTrackers {
		req.Msg.Series = append(req.Msg.Series, p.profile)
	}

	_, err = c.(PushClient).Push(ctx, req)
	return err
}

type profileTracker struct {
	profile     *pushv1.RawProfileSeries
	minSuccess  int
	maxFailures int
	succeeded   atomic.Int32
	failed      atomic.Int32
}

type pushTracker struct {
	samplesPending atomic.Int32
	samplesFailed  atomic.Int32
	done           chan struct{}
	err            chan error
}

func labelsString(ls []*commonv1.LabelPair) string {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, l := range ls {
		if i > 0 {
			b.WriteByte(',')
			b.WriteByte(' ')
		}
		b.WriteString(l.Name)
		b.WriteByte('=')
		b.WriteString(strconv.Quote(l.Value))
	}
	b.WriteByte('}')
	return b.String()
}

// TokenFor generates a token used for finding ingesters from ring
func TokenFor(tenantID, labels string) uint32 {
	h := fnv.New32()
	_, _ = h.Write([]byte(tenantID))
	_, _ = h.Write([]byte(labels))
	return h.Sum32()
}

func (d *Distributor) ParcaProfileStore() parcastorev1.ProfileStoreServiceServer {
	return &ParcaProfileStore{
		distributor: d,
	}
}

type ParcaProfileStore struct {
	parcastorev1.UnimplementedProfileStoreServiceServer
	distributor *Distributor
}

func (s *ParcaProfileStore) WriteRaw(ctx context.Context, req *parcastorev1.WriteRawRequest) (*parcastorev1.WriteRawResponse, error) {
	nReq := &pushv1.PushRequest{
		Series: make([]*pushv1.RawProfileSeries, len(req.Series)),
	}
	for idxSeries, series := range req.Series {
		nReq.Series[idxSeries] = &pushv1.RawProfileSeries{
			Samples: make([]*pushv1.RawSample, len(series.Samples)),
			Labels:  make([]*commonv1.LabelPair, len(series.Labels.Labels)),
		}
		for idx, l := range series.Labels.Labels {
			nReq.Series[idxSeries].Labels[idx] = &commonv1.LabelPair{
				Name:  l.Name,
				Value: l.Value,
			}
		}
		for idx, s := range series.Samples {
			nReq.Series[idxSeries].Samples[idx] = &pushv1.RawSample{
				RawProfile: s.RawProfile,
			}
		}
		level.Warn(s.distributor.logger).Log("msg", "converted parca sample", "labels", firemodel.LabelPairsString(nReq.Series[idxSeries].Labels))
	}

	if _, err := s.distributor.Push(ctx, connect.NewRequest(nReq)); err != nil {
		return nil, err
	}

	return &parcastorev1.WriteRawResponse{}, nil
}
