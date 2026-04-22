package service

import (
	"context"
	"time"

	keygen "github.com/Zapi-web/gopher-pinger/internal/keyGen"
	"github.com/Zapi-web/gopher-pinger/internal/pinger"
	"github.com/oklog/ulid/v2"
)

type PingerService interface {
	StartMonitoring(url string, interval int) (ulid.ULID, error)
}

type ProcessStore interface {
	Set(key ulid.ULID, value context.CancelFunc) error
	Get(key ulid.ULID) (context.CancelFunc, error)
	Delete(key ulid.ULID) error
}

type pingerService struct {
	processes ProcessStore
}

func NewService(p ProcessStore) PingerService {
	return &pingerService{
		processes: p,
	}
}

func (s *pingerService) StartMonitoring(url string, interval int) (ulid.ULID, error) {
	id, err := keygen.GetKey()
	if err != nil {
		return ulid.ULID{}, err
	}

	cancel := pinger.Start(url, time.Duration(interval)*time.Second)
	err = s.processes.Set(id, cancel)

	if err != nil {
		return ulid.ULID{}, err
	}

	return id, nil
}
