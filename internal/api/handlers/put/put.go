package put

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Zapi-web/gopher-pinger/internal/api/handlers/utils"
	"github.com/Zapi-web/gopher-pinger/internal/domain"
	"github.com/Zapi-web/gopher-pinger/internal/service"
)

type Request struct {
	ID       string `json:"id"`
	Interval int    `json:"interval"`
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

		err = pinger.UpdateProcess(r.Context(), req.ID, req.Interval)

		if err != nil {
			if errors.Is(err, domain.ErrInvalidId) || errors.Is(err, domain.ErrInputisEmpty) {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			if errors.Is(err, domain.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			slog.Error("Failed to update data in database", "err", err, "ULID", req.ID)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		slog.Info("new interval setted", "ULID", req.ID, "interval", req.Interval)

		w.WriteHeader(http.StatusNoContent)
	}
}
