package phlaredb

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	phlarecontext "github.com/grafana/phlare/pkg/phlare/context"
	query "github.com/grafana/phlare/pkg/phlaredb/query"
)

type contextKey uint8

const (
	headMetricsContextKey contextKey = iota
	blockMetricsContextKey
)

type headMetrics struct {
	head *Head

	series        prometheus.GaugeFunc
	seriesCreated *prometheus.CounterVec

	profiles        prometheus.GaugeFunc
	profilesCreated *prometheus.CounterVec

	sizeBytes   *prometheus.GaugeVec
	rowsWritten *prometheus.CounterVec

	sampleValuesIngested *prometheus.CounterVec
	sampleValuesReceived *prometheus.CounterVec
}

func newHeadMetrics(reg prometheus.Registerer) *headMetrics {
	m := &headMetrics{
		seriesCreated: promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "phlare_tsdb_head_series_created_total",
			Help: "Total number of series created in the head",
		}, []string{"profile_name"}),
		rowsWritten: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "phlare_rows_written",
				Help: "Number of rows written to a parquet table.",
			},
			[]string{"type"}),
		profilesCreated: promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "phlare_head_profiles_created_total",
			Help: "Total number of profiles created in the head",
		}, []string{"profile_name"}),
		sampleValuesIngested: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "phlare_head_ingested_sample_values_total",
				Help: "Number of sample values ingested into the head per profile type.",
			},
			[]string{"profile_name"}),
		sampleValuesReceived: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "phlare_head_received_sample_values_total",
				Help: "Number of sample values received into the head per profile type.",
			},
			[]string{"profile_name"}),

		// this metric is not registered using promauto, as it has a callback into the header
		sizeBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "phlare_head_size_bytes",
				Help: "Size of a particular in memory store within the head phlaredb block.",
			},
			[]string{"type"}),
	}

	// metrics that call into the head
	m.series = promauto.With(reg).NewGaugeFunc(prometheus.GaugeOpts{
		Name: "phlare_tsdb_head_series",
		Help: "Total number of series in the head block.",
	}, func() float64 {
		if m.head == nil {
			return 0.0
		}
		return float64(m.head.profiles.index.totalSeries.Load())
	})
	m.profiles = promauto.With(reg).NewGaugeFunc(prometheus.GaugeOpts{
		Name: "phlare_head_profiles",
		Help: "Total number of profiles in the head block.",
	}, func() float64 {
		if m.head == nil {
			return 0.0
		}
		return float64(m.head.profiles.index.totalProfiles.Load())
	})

	if reg != nil {
		reg.MustRegister(
			m,
		)
	}
	return m
}

func contextWithHeadMetrics(ctx context.Context, m *headMetrics) context.Context {
	return context.WithValue(ctx, headMetricsContextKey, m)
}

func contextHeadMetrics(ctx context.Context) *headMetrics {
	m, ok := ctx.Value(headMetricsContextKey).(*headMetrics)
	if !ok {
		return newHeadMetrics(phlarecontext.Registry(ctx))
	}
	return m
}

func (m *headMetrics) setHead(head *Head) *headMetrics {
	m.head = head
	return m
}

func (m *headMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.sizeBytes.Describe(ch)
}

func (m *headMetrics) Collect(ch chan<- prometheus.Metric) {
	if m.head != nil {
		for _, t := range m.head.tables {
			m.sizeBytes.WithLabelValues(t.Name()).Set(float64(t.Size()))
		}
	}
	m.sizeBytes.Collect(ch)
}

type blocksMetrics struct {
	query *query.Metrics

	blockOpeningLatency prometheus.Histogram
}

func newBlocksMetrics(reg prometheus.Registerer) *blocksMetrics {
	return &blocksMetrics{
		query: query.NewMetrics(reg),
		blockOpeningLatency: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name: "phlaredb_block_opening_duration",
			Help: "Latency of opening a block in seconds",
		}),
	}
}

func contextWithBlockMetrics(ctx context.Context, m *blocksMetrics) context.Context {
	return context.WithValue(ctx, blockMetricsContextKey, m)
}

func contextBlockMetrics(ctx context.Context) *blocksMetrics {
	m, ok := ctx.Value(blockMetricsContextKey).(*blocksMetrics)
	if !ok {
		return newBlocksMetrics(phlarecontext.Registry(ctx))
	}
	return m
}
