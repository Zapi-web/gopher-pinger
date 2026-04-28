package post

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Zapi-web/gopher-pinger/internal/api/handlers/utils"
	"github.com/Zapi-web/gopher-pinger/internal/service"
)

type Request struct {
	URL      string `json:"url"`
	Interval int    `json:"interval"`
}

type Response struct {
	Id string `json:"id"`
}

func New(pinger service.PingerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		r.Body = http.MaxBytesReader(w, r.Body, 1024*10)

		req, err := utils.Decode[Request](r)

		if err != nil {
			slog.Info("failed to decode request body", "err", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		id, err := pinger.StartMonitoring(r.Context(), req.URL, req.Interval)
		if err != nil {
			slog.Error("failed to start monitoring", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		slog.Info("monitoring started", "ulid", id)

		res := Response{
			Id: id.String(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Warn("failed to encode response", "err", err)
		}
	}
}
