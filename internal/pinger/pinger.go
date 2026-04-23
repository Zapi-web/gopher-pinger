package pinger

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type CheckResult struct {
	ID     string
	Status int
}

var sharedClient = &http.Client{
	Timeout: 10 * time.Second,
}

func Start(id string, url string, interval time.Duration, results chan<- CheckResult) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("pinger stopped", "ULID", id, "url", url)
				return
			case <-ticker.C:
				status, err := ping(ctx, url)

				if err != nil {
					slog.Warn("ping error", "url", url, "err", err)
					continue
				}

				results <- CheckResult{ID: id, Status: status}

				slog.Info("url pinged", "ULID", id, "url", url, "code", status)
			}
		}
	}()

	slog.Info("pinger started", "ULID", id, "url", url, "interval", interval)

	return cancel
}

func ping(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := sharedClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
