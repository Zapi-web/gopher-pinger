package local

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Zapi-web/gopher-pinger/internal/domain"
	"github.com/oklog/ulid/v2"
)

type MapStorage struct {
	mu     sync.RWMutex
	mapStr map[ulid.ULID]context.CancelFunc
}

func InitMap() *MapStorage {
	newMap := MapStorage{
		mapStr: make(map[ulid.ULID]context.CancelFunc),
	}

	return &newMap
}

func (m *MapStorage) Set(key ulid.ULID, value context.CancelFunc) error {
	if key == (ulid.ULID{}) || value == nil {
		return domain.ErrInputisEmpty
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.mapStr[key] = value
	slog.Info("key added", "ULID", key)

	return nil
}

func (m *MapStorage) Get(key ulid.ULID) (context.CancelFunc, error) {
	if key == (ulid.ULID{}) {
		return nil, domain.ErrInputisEmpty
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.mapStr[key]

	if !ok {
		return nil, fmt.Errorf("given key is not exist %s", key)
	}

	return value, nil
}

func (m *MapStorage) Delete(key ulid.ULID) error {
	if key == (ulid.ULID{}) {
		return domain.ErrInputisEmpty
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.mapStr[key]; !ok {
		return fmt.Errorf("given key is not exist %s", key)
	}

	delete(m.mapStr, key)
	slog.Info("key deleted", "ULID", key)

	return nil
}
