package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Zapi-web/gopher-pinger/internal/domain"
	keygen "github.com/Zapi-web/gopher-pinger/internal/keyGen"
	"github.com/Zapi-web/gopher-pinger/internal/pinger"
	"github.com/oklog/ulid/v2"
)

type PingerService interface {
	StartMonitoring(ctx context.Context, url string, interval int) (ulid.ULID, error)
	GetProcess(ctx context.Context, id string) (Target, error)
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
}

func NewService(p ProcessStore, s StateStore) PingerService {
	return &pingerService{
		processes: p,
		state:     s,
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
		return ulid.ULID{}, fmt.Errorf("failed setting data in database: %w", err)
	}

	err = s.state.Set(ctx, id.String(), Target{
		URL:      url,
		Interval: interval,
	})

	if err != nil {
		return ulid.ULID{}, err
	}

	return id, nil
}

func (s *pingerService) GetProcess(ctx context.Context, id string) (Target, error) {
	var data Target

	if id == "" {
		return data, domain.ErrInputisEmpty
	}

	data, err := s.state.Get(ctx, id)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return data, err
		}

		return data, fmt.Errorf("failed getting data from database: %w", err)
	}

	return data, nil
}

func (s *pingerService) ResultsMonitoring() {
	for res := range s.results {
		s.state.UpdateStatus(context.Background(), res.ID, res.Status, time.Now().Format(time.RFC3339))
	}
}
