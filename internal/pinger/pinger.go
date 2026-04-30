package pinger

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type CheckResult struct {
	ID       string
	URL      string
	Status   int
	Duration time.Duration
}

var sharedClient = &http.Client{
	Timeout: 10 * time.Second,
}

type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}

func Start(ctx context.Context, locker Locker, id string, url string, interval time.Duration, results chan<- CheckResult) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("pinger stopped", "ulid", id, "url", url)
				return
			case <-ticker.C:
				start := time.Now()

				ok, err := locker.Lock(ctx, id, interval-(time.Microsecond*100))

				if err != nil {
					slog.Error("failed to lock", "ulid", id, "err", err)
					continue
				}

				if !ok {
					continue
				}

				status, err := ping(ctx, url)
				dur := time.Since(start)

				if err != nil {
					slog.Warn("ping error", "url", url, "err", err)
					status = -1
				}

				select {
				case results <- CheckResult{ID: id, URL: url, Status: status, Duration: dur}:
					slog.Info("url pinged", "ulid", id, "url", url, "code", status)
				default:
					slog.Warn("results channel is full", "ulid", id)
				}
			}
		}
	}()

	slog.Info("pinger started", "ulid", id, "url", url, "interval", interval)

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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "err", err)
		}
	}()

	return resp.StatusCode, nil
}
