package get

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Zapi-web/gopher-pinger/internal/api/handlers/utils"
	"github.com/Zapi-web/gopher-pinger/internal/domain"
	"github.com/Zapi-web/gopher-pinger/internal/service"
)

type Request struct {
	ID string `json:"id"`
}

type Response struct {
	URL             string `json:"url"`
	LastTimeChecked string `json:"last_time_checked"`
	LastCode        int    `json:"last_code"`
	Interval        int    `json:"interval"`
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

		data, err := pinger.GetProcess(r.Context(), req.ID)

		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			if errors.Is(err, domain.ErrInvalidId) || errors.Is(err, domain.ErrInputisEmpty) {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}

			slog.Warn("failed to get data from database", "ULID", req.ID, "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		res := Response{
			URL:             data.URL,
			LastTimeChecked: data.LastTimeChecked,
			LastCode:        data.LastCode,
			Interval:        data.Interval,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Warn("failed to encode response", "err", err)
		}
	}
}
