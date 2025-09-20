package observability

import (
    "net/http"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    EventsTotal     = prometheus.NewCounter(prometheus.CounterOpts{Name: "server_analyst_events_total", Help: "Total events ingested"})
    DetectionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "server_analyst_detections_total", Help: "Total detections"}, []string{"category", "rule"})
    DecisionsTotal  = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "server_analyst_decisions_total", Help: "Total decisions"}, []string{"action"})
    UploadsTotal    = prometheus.NewCounter(prometheus.CounterOpts{Name: "server_analyst_uploads_total", Help: "Total uploads to MinIO"})
    UploadsFailed   = prometheus.NewCounter(prometheus.CounterOpts{Name: "server_analyst_uploads_failed_total", Help: "Failed uploads to MinIO"})
    SpoolQueueSize  = prometheus.NewGauge(prometheus.GaugeOpts{Name: "server_analyst_spool_queue_size", Help: "Current spool queue size"})
)

func InitMetrics() {
    prometheus.MustRegister(EventsTotal, DetectionsTotal, DecisionsTotal, UploadsTotal, UploadsFailed, SpoolQueueSize)
}

func MetricsHandler() http.Handler { return promhttp.Handler() }
