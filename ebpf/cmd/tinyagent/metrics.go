package main

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	targetsActive                 prometheus.Gauge
	profilingSessionsTotal        prometheus.Counter
	profilingSessionsFailingTotal prometheus.Counter
	pprofsTotal                   prometheus.Counter
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		targetsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "pyroscope_ebpf_active_targets",
			Help: "Current number of active targets being tracked by the ebpf component",
		}),
		profilingSessionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_profiling_sessions_total",
			Help: "Total number of profiling sessions started by the ebpf component",
		}),
		profilingSessionsFailingTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_profiling_sessions_failing_total",
			Help: "Total number of profiling sessions failed to complete by the ebpf component",
		}),
		pprofsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pyroscope_ebpf_pprofs_total",
			Help: "Total number of pprof profiles collected by the ebpf component",
		}),
	}

	if reg != nil {
		reg.MustRegister(
			m.targetsActive,
			m.profilingSessionsTotal,
			m.profilingSessionsFailingTotal,
			m.pprofsTotal,
		)
	}

	return m
}
