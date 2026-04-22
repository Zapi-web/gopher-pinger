package pinger

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

var sharedClient = &http.Client{
	Timeout: 10 * time.Second,
}

func Start(url string, interval time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("pinger stopped", "url", url)
				return
			case <-ticker.C:
				status, err := ping(ctx, url)

				if err != nil {
					slog.Warn("ping error", "url", url, "err", err)
					continue
				}

				slog.Info("url pinged", "url", url, "code", status)
			}
		}
	}()

	slog.Info("pinger started", "url", url, "interval", interval)

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
