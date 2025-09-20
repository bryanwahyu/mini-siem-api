package metrics

import "github.com/prometheus/client_golang/prometheus"

// Metrics exposes Prometheus counters for the mini SIEM API.
type Metrics struct {
	registry        prometheus.Registerer
	EventsTotal     prometheus.Counter
	DetectionsTotal *prometheus.CounterVec
	ErrorsTotal     prometheus.Counter
}

// New initialises the metrics collectors and registers them.
func New(reg prometheus.Registerer) *Metrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	m := &Metrics{
		registry: reg,
		EventsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mini_siem_events_total",
			Help: "Total number of events ingested",
		}),
		DetectionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "mini_siem_detections_total",
			Help: "Detections emitted by severity",
		}, []string{"severity"}),
		ErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "mini_siem_errors_total",
			Help: "Total number of processing errors",
		}),
	}

	reg.MustRegister(m.EventsTotal, m.DetectionsTotal, m.ErrorsTotal)
	return m
}

// RecordDetection increments detection counter for severity.
func (m *Metrics) RecordDetection(severity string) {
	if m == nil {
		return
	}
	m.DetectionsTotal.WithLabelValues(severity).Inc()
}

// RecordEvent increments event counter.
func (m *Metrics) RecordEvent() {
	if m == nil {
		return
	}
	m.EventsTotal.Inc()
}

// RecordError increments error counter.
func (m *Metrics) RecordError() {
	if m == nil {
		return
	}
	m.ErrorsTotal.Inc()
}
