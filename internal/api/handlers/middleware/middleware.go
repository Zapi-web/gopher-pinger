package logMiddleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Zapi-web/gopher-pinger/internal/service"
)

type ResponsWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *ResponsWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler, metrics service.PingerMetrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &ResponsWriter{w, http.StatusOK}
		slog.Info("new request", "method", r.Method)

		timeNow := time.Now()
		next.ServeHTTP(rw, r)

		duration := time.Since(timeNow)
		metrics.NewRequest(r.Method, rw.statusCode, duration)
	})
}
