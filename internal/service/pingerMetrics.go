package service

import (
	"strconv"
	"time"

	"github.com/Zapi-web/gopher-pinger/internal/metrics"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=PingerMetrics
type PingerMetrics interface {
	NewRequest(method string, status int, duration time.Duration)
	NewPing(URL string, status int, duration time.Duration)
	IncWorker()
	DecWorker()
}

type metricService struct {
	metrics metrics.Metrics
}

func NewMetricsService(m metrics.Metrics) *metricService {
	return &metricService{
		metrics: m,
	}
}

func (m *metricService) NewRequest(method string, status int, duration time.Duration) {
	strStatus := strconv.Itoa(status)

	m.metrics.RequestsTotal.WithLabelValues(method, strStatus).Inc()
	m.metrics.RequestDuration.WithLabelValues(method, strStatus).Observe(duration.Seconds())
}

func (m *metricService) NewPing(URL string, status int, duration time.Duration) {
	strStatus := strconv.Itoa(status)

	m.metrics.PingsTotal.WithLabelValues(URL, strStatus).Inc()
	m.metrics.PingDuration.WithLabelValues(URL).Observe(duration.Seconds())
}

func (m *metricService) IncWorker() {
	m.metrics.ActiveWorkers.Inc()
}

func (m *metricService) DecWorker() {
	m.metrics.ActiveWorkers.Dec()
}
