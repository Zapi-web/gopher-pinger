package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Zapi-web/gopher-pinger/internal/domain"
	keygen "github.com/Zapi-web/gopher-pinger/internal/keyGen"
	"github.com/Zapi-web/gopher-pinger/internal/pinger"
	"github.com/oklog/ulid/v2"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=PingerService
type PingerService interface {
	StartMonitoring(ctx context.Context, url string, interval int) (ulid.ULID, error)
	GetProcess(ctx context.Context, id string) (domain.Target, error)
	DeleteProcess(ctx context.Context, id string) error
	UpdateProcess(ctx context.Context, id string, interval int) error
	ResultsMonitoring()
	Init() error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=ProcessStore
type ProcessStore interface {
	Set(key ulid.ULID, value context.CancelFunc) error
	Get(key ulid.ULID) (context.CancelFunc, error)
	Delete(key ulid.ULID) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=StateStore
type StateStore interface {
	Set(ctx context.Context, key string, value domain.Target) error
	Get(ctx context.Context, key string) (domain.Target, error)
	Delete(ctx context.Context, key string) error
	UpdateStatus(ctx context.Context, key string, code int, timestamp string) error
	GetAll(ctx context.Context) ([]domain.Target, error)
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}

type pingerService struct {
	processes ProcessStore
	state     StateStore
	results   chan pinger.CheckResult
	metrics   PingerMetrics
	appCtx    context.Context
}

func NewService(ctx context.Context, p ProcessStore, s StateStore, m PingerMetrics) PingerService {
	return &pingerService{
		processes: p,
		state:     s,
		results:   make(chan pinger.CheckResult, 100),
		metrics:   m,
		appCtx:    ctx,
	}
}

func (s *pingerService) Init() error {
	now := time.Now()
	targets, err := s.state.GetAll(s.appCtx)

	if err != nil {
		return fmt.Errorf("failed to get array of targets: %w", err)
	}

	for _, target := range targets {
		ulid, err := ulid.Parse(target.ID)

		if err != nil {
			slog.Warn("bad id in initialize", "ulid", ulid)
			continue
		}

		cancel := pinger.Start(s.appCtx, s.state, target.ID, target.URL, time.Duration(target.Interval)*time.Second, s.results)

		err = s.processes.Set(ulid, cancel)

		if err != nil {
			cancel()
			slog.Error("failed to set key in map", "ulid", ulid, "err", err)
			continue
		}
	}

	dur := time.Since(now)
	slog.Info("init complete", "dur", dur.String())

	return nil
}

func (s *pingerService) StartMonitoring(reqCtx context.Context, url string, interval int) (ulid.ULID, error) {
	if interval <= 0 {
		return ulid.ULID{}, domain.ErrInvalidInterval
	}

	id, err := keygen.GetKey()
	if err != nil {
		return ulid.ULID{}, err
	}

	cancel := pinger.Start(s.appCtx, s.state, id.String(), url, time.Duration(interval)*time.Second, s.results)
	err = s.processes.Set(id, cancel)

	if err != nil {
		cancel()
		return ulid.ULID{}, fmt.Errorf("failed setting data in map: %w", err)
	}

	err = s.state.Set(reqCtx, id.String(), domain.Target{
		ID:       id.String(),
		URL:      url,
		Interval: interval,
	})

	if err != nil {
		cancel()
		s.processes.Delete(id)
		return ulid.ULID{}, fmt.Errorf("failed setting data in database: %w", err)
	}

	s.metrics.IncWorker()
	return id, nil
}

func (s *pingerService) GetProcess(ctx context.Context, id string) (domain.Target, error) {
	var data domain.Target

	if id == "" {
		return data, domain.ErrInputisEmpty
	}

	_, err := ulid.Parse(id)

	if err != nil {
		return data, domain.ErrInvalidId
	}

	data, err = s.state.Get(ctx, id)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return data, err
		}

		return data, fmt.Errorf("failed getting data from database: %w", err)
	}

	return data, nil
}

func (s *pingerService) DeleteProcess(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInputisEmpty
	}

	ulid, err := ulid.Parse(id)

	if err != nil {
		return domain.ErrInvalidId
	}

	err = s.state.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	cancel, err := s.processes.Get(ulid)
	if err != nil {
		slog.Warn("not found process after deleting from state", "ULID", id)
		return nil
	}

	cancel()
	err = s.processes.Delete(ulid)

	if err != nil {
		return fmt.Errorf("failed to remove from map: %w", err)
	}

	s.metrics.DecWorker()
	return nil
}

func (s *pingerService) UpdateProcess(reqCtx context.Context, id string, interval int) error {
	if interval <= 0 {
		return domain.ErrInvalidInterval
	}

	if id == "" {
		return domain.ErrInputisEmpty
	}

	ulid, err := ulid.Parse(id)

	if err != nil {
		return domain.ErrInvalidId
	}

	data, err := s.GetProcess(reqCtx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}

		return fmt.Errorf("failed getting old data: %w", err)
	}

	err = s.DeleteProcess(reqCtx, id)
	if err != nil {
		return fmt.Errorf("failed to delete old process: %w", err)
	}

	cancel := pinger.Start(s.appCtx, s.state, id, data.URL, time.Duration(interval)*time.Second, s.results)
	err = s.processes.Set(ulid, cancel)

	if err != nil {
		cancel()
		return fmt.Errorf("failed to set new process in map: %w", err)
	}

	data.Interval = interval
	err = s.state.Set(reqCtx, id, data)

	if err != nil {
		cancel()
		return fmt.Errorf("failed set a data in database: %w", err)
	}

	s.metrics.IncWorker()
	return nil
}

func (s *pingerService) ResultsMonitoring() {
	for {
		select {
		case res, ok := <-s.results:
			if !ok {
				slog.Info("results channel closed, shutting down result monitoring")
				return
			}

			err := s.state.UpdateStatus(s.appCtx, res.ID, res.Status, time.Now().Format(time.RFC3339))
			if err != nil {
				slog.Error("failed update status", "ulid", res.ID, "err", err)
			}

			s.metrics.NewPing(res.URL, res.Status, res.Duration)
		case <-s.appCtx.Done():
			slog.Info("stopping result monitoring", "reason", s.appCtx.Err())
			return
		}
	}
}
