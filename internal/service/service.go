package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Zapi-web/gopher-pinger/internal/domain"
	keygen "github.com/Zapi-web/gopher-pinger/internal/keyGen"
	"github.com/Zapi-web/gopher-pinger/internal/metrics"
	"github.com/Zapi-web/gopher-pinger/internal/pinger"
	"github.com/oklog/ulid/v2"
)

type PingerService interface {
	StartMonitoring(ctx context.Context, url string, interval int) (ulid.ULID, error)
	GetProcess(ctx context.Context, id string) (Target, error)
	DeleteProcess(ctx context.Context, id string) error
	UpdateProcess(ctx context.Context, id string, interval int) error
}

type ProcessStore interface {
	Set(key ulid.ULID, value context.CancelFunc) error
	Get(key ulid.ULID) (context.CancelFunc, error)
	Delete(key ulid.ULID) error
}

type StateStore interface {
	Set(ctx context.Context, key string, value Target) error
	Get(ctx context.Context, key string) (Target, error)
	Delete(ctx context.Context, key string) error
	UpdateStatus(ctx context.Context, key string, code int, timestamp string) error
}

type Target struct {
	URL             string
	LastTimeChecked string
	LastCode        int
	Interval        int
}

type pingerService struct {
	processes ProcessStore
	state     StateStore
	results   chan pinger.CheckResult
	metrics   *metrics.Metrics
}

func NewService(p ProcessStore, s StateStore, m *metrics.Metrics) PingerService {
	return &pingerService{
		processes: p,
		state:     s,
		results:   make(chan pinger.CheckResult, 100),
		metrics:   m,
	}
}

func (s *pingerService) StartMonitoring(ctx context.Context, url string, interval int) (ulid.ULID, error) {
	id, err := keygen.GetKey()
	if err != nil {
		return ulid.ULID{}, err
	}

	cancel := pinger.Start(id.String(), url, time.Duration(interval)*time.Second, s.results)
	err = s.processes.Set(id, cancel)

	if err != nil {
		cancel()
		return ulid.ULID{}, fmt.Errorf("failed setting data in map: %w", err)
	}

	err = s.state.Set(ctx, id.String(), Target{
		URL:      url,
		Interval: interval,
	})

	if err != nil {
		cancel()
		s.processes.Delete(id)
		return ulid.ULID{}, fmt.Errorf("failed setting data in database: %w", err)
	}

	s.metrics.ActiveWorkers.Inc()
	return id, nil
}

func (s *pingerService) GetProcess(ctx context.Context, id string) (Target, error) {
	var data Target

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
		return nil
	}

	cancel()
	err = s.processes.Delete(ulid)

	if err != nil {
		return fmt.Errorf("failed to remove from map: %w", err)
	}

	s.metrics.ActiveWorkers.Dec()
	return nil
}

func (s *pingerService) UpdateProcess(ctx context.Context, id string, interval int) error {
	if id == "" {
		return domain.ErrInputisEmpty
	}

	ulid, err := ulid.Parse(id)

	if err != nil {
		return domain.ErrInvalidId
	}

	data, err := s.GetProcess(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}

		return fmt.Errorf("failed getting old data: %w", err)
	}

	err = s.DeleteProcess(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete old process: %w", err)
	}

	cancel := pinger.Start(id, data.URL, time.Duration(interval)*time.Second, s.results)
	err = s.processes.Set(ulid, cancel)

	if err != nil {
		cancel()
		return fmt.Errorf("failed to set new process in map: %w", err)
	}

	data.Interval = interval
	err = s.state.Set(ctx, id, data)

	if err != nil {
		cancel()
		return fmt.Errorf("failed set a data in database: %w", err)
	}

	return nil
}

func (s *pingerService) ResultsMonitoring() {
	for res := range s.results {
		s.state.UpdateStatus(context.Background(), res.ID, res.Status, time.Now().Format(time.RFC3339))
	}
}
