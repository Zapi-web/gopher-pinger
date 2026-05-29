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
	"github.com/Zapi-web/gopher-pinger/internal/service/utils"
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
	Set(key ulid.ULID, value *domain.ActiveProcess) error
	Get(key ulid.ULID) (*domain.ActiveProcess, error)
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

		gorData := pinger.GoroutineData{
			ID:       target.ID,
			URL:      target.URL,
			Interval: time.Duration(target.Interval) * time.Second,
			Results:  s.results,
		}

		cancel, ticker := pinger.Start(s.appCtx, s.state, &gorData)

		err = s.processes.Set(ulid, &domain.ActiveProcess{Cancel: cancel, Ticker: ticker})

		if err != nil {
			cancel()
			ticker.Stop()
			slog.Error("failed to set key in map", "ulid", ulid, "err", err)
			continue
		}
	}

	dur := time.Since(now)
	slog.Info("init complete", "dur", dur.String())

	return nil
}

func (s *pingerService) StartMonitoring(reqCtx context.Context, reqUrl string, interval int) (ulid.ULID, error) {
	if interval <= 0 {
		return ulid.ULID{}, domain.ErrInvalidInterval
	}

	if !utils.IsValidURL(reqUrl) {
		return ulid.ULID{}, domain.ErrInvalidURL
	}

	id, err := keygen.GetKey()
	if err != nil {
		return ulid.ULID{}, err
	}

	gorData := pinger.GoroutineData{
		ID:       id.String(),
		URL:      reqUrl,
		Interval: time.Duration(interval) * time.Second,
		Results:  s.results,
	}

	cancel, ticker := pinger.Start(s.appCtx, s.state, &gorData)
	err = s.processes.Set(id, &domain.ActiveProcess{Cancel: cancel, Ticker: ticker})

	if err != nil {
		cancel()
		ticker.Stop()
		return ulid.ULID{}, fmt.Errorf("failed setting data in map: %w", err)
	}

	err = s.state.Set(reqCtx, id.String(), domain.Target{
		ID:       id.String(),
		URL:      reqUrl,
		Interval: interval,
	})

	if err != nil {
		cancel()
		ticker.Stop()
		mapErr := s.processes.Delete(id)

		if !errors.Is(mapErr, domain.ErrNotFound) {
			slog.Warn("failed deleting a data from map", "ulid", id, "err", err)
		}

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

	proc, err := s.processes.Get(ulid)
	if err != nil {
		slog.Warn("not found process after deleting from state", "ULID", id)
		return nil
	}

	proc.Cancel()
	proc.Ticker.Stop()
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

	tar, err := s.state.Get(reqCtx, id)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}

		return fmt.Errorf("failed to get old data in database %w", err)
	}

	tar.Interval = interval

	proc, err := s.processes.Get(ulid)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}

		return fmt.Errorf("failed to get old process %w", err)
	}

	err = s.state.Set(reqCtx, id, tar)

	if err != nil {
		return fmt.Errorf("failed to set new interval in database %w", err)
	}

	proc.Ticker.Reset(time.Duration(interval) * time.Second)

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
