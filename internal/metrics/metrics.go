package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	PingsTotal      *prometheus.CounterVec
	PingDuration    *prometheus.HistogramVec
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ActiveWorkers   prometheus.Gauge
}

func New() *Metrics {
	return &Metrics{
		PingsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "pinger_pings_total",
			Help: "Total number of pings",
		}, []string{"url", "status"}),

		PingDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "pinger_ping_duration_seconds",
			Help:    "Time taken to ping a URL",
			Buckets: prometheus.DefBuckets,
		}, []string{"url"}),

		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "pinger_total_requests",
			Help: "Total number of requests",
		}, []string{"method", "status"}),

		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "pinger_request_duration_seconds",
			Help:    "Time taken to ping a URL",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "status"}),

		ActiveWorkers: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "pinger_active_workers_count",
			Help: "Number of currently active goroutines",
		}),
	}
}
