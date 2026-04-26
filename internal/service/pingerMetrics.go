package service

import (
	"strconv"
	"time"

	"github.com/Zapi-web/gopher-pinger/internal/metrics"
)

type PingerMetrics interface {
	NewRequest(method string, status int, duration time.Duration)
}

type metricService struct {
	metrics *metrics.Metrics
}

func (m *metricService) NewRequest(method string, status int, duration time.Duration) {
	strStatus := strconv.Itoa(status)

	m.metrics.RequestsTotal.WithLabelValues(method, strStatus).Inc()
	m.metrics.RequestDuration.WithLabelValues(method, strStatus).Observe(duration.Seconds())
}
