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

type GoroutineData struct {
	ID       string
	URL      string
	Interval time.Duration
	Results  chan<- CheckResult
}

var sharedClient = &http.Client{
	Timeout: 10 * time.Second,
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=Locker
type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}

func Start(ctx context.Context, locker Locker, req *GoroutineData) (context.CancelFunc, *time.Ticker) {
	ctx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(req.Interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("pinger stopped", "ulid", req.ID, "url", req.URL)
				return
			case <-ticker.C:

				ok, err := locker.Lock(ctx, req.ID, req.Interval-(time.Microsecond*100))

				if err != nil {
					slog.Error("failed to lock", "ulid", req.ID, "err", err)
					continue
				}

				if !ok {
					continue
				}

				start := time.Now()
				status, err := ping(ctx, req.URL)
				dur := time.Since(start)

				if err != nil {
					slog.Warn("ping error", "url", req.URL, "err", err)
					status = -1
				}

				select {
				case req.Results <- CheckResult{ID: req.ID, URL: req.URL, Status: status, Duration: dur}:
					slog.Info("url pinged", "ulid", req.ID, "url", req.URL, "code", status)
				default:
					slog.Warn("results channel is full", "ulid", req.ID)
				}
			}
		}
	}()

	slog.Info("pinger started", "ulid", req.ID, "url", req.URL, "interval", req.Interval)

	return cancel, ticker
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
