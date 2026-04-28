package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Zapi-web/gopher-pinger/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisDb struct {
	rdb *redis.Client
}

type dataStruct struct {
	ID              string `redis:"id"`
	Url             string `redis:"url"`
	LastTimeChecked string `redis:"last_time_checked"`
	LastCode        int    `redis:"last_code"`
	Interval        int    `redis:"interval"`
}

func New(addr string) (*RedisDb, error) {
	var r RedisDb

	r.rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	err := r.rdb.Ping(context.Background()).Err()

	if err != nil {
		return nil, fmt.Errorf("failed to ping a database %w", err)
	}

	return &r, nil
}

func (r *RedisDb) Set(ctx context.Context, key string, value domain.Target) error {
	if key == "" || value.URL == "" {
		return domain.ErrInputisEmpty
	}

	setValue := dataStruct{
		ID:              value.ID,
		Url:             value.URL,
		LastTimeChecked: value.LastTimeChecked,
		LastCode:        value.LastCode,
		Interval:        value.Interval,
	}

	err := r.rdb.HSet(ctx, key, setValue).Err()

	if err != nil {
		return fmt.Errorf("failed to set a value to database %w", err)
	}

	slog.Info("key added to redis", "ULID", key)

	return nil
}

func (r *RedisDb) Get(ctx context.Context, key string) (domain.Target, error) {
	if key == "" {
		return domain.Target{}, domain.ErrInputisEmpty
	}
	var res dataStruct

	err := r.rdb.HGetAll(ctx, key).Scan(&res)

	if err != nil {
		return domain.Target{}, fmt.Errorf("failed to get a value from a database %w", err)
	}

	if res.Url == "" {
		return domain.Target{}, domain.ErrNotFound
	}

	val := domain.Target{
		ID:              res.ID,
		URL:             res.Url,
		LastTimeChecked: res.LastTimeChecked,
		LastCode:        res.LastCode,
		Interval:        res.Interval,
	}

	slog.Info("value retrieved from redis", "ULID", key)

	return val, nil
}

func (r *RedisDb) Delete(ctx context.Context, key string) error {
	if key == "" {
		return domain.ErrInputisEmpty
	}

	err := r.rdb.Del(ctx, key).Err()

	if err != nil {
		return fmt.Errorf("failed to delete a key from database %w", err)
	}

	slog.Info("key deleted from redis", "ULID", key)

	return nil
}

func (r *RedisDb) Close() {
	_ = r.rdb.Close()
}

func (r *RedisDb) UpdateStatus(ctx context.Context, key string, code int, timestamp string) error {
	ok, err := r.rdb.HExists(ctx, key, "url").Result()

	if err != nil {
		return fmt.Errorf("failed to check is key exist: %w", err)
	}

	if !ok {
		return domain.ErrNotFound
	}

	err = r.rdb.HSet(ctx, key,
		"last_code", code,
		"last_time_checked", timestamp,
	).Err()

	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (r *RedisDb) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	err := r.rdb.SetArgs(ctx, "lock:"+key, "busy", redis.SetArgs{
		TTL:  ttl,
		Mode: "NX",
	}).Err()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return false, fmt.Errorf("failed to lock db key")
	}

	return true, nil
}

func (r *RedisDb) Unlock(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, "lock:"+key).Err()
}

func (r *RedisDb) GetAll(ctx context.Context) ([]domain.Target, error) {
	var cursor uint64
	var totalTargets []domain.Target

	for {
		keys, nextCursor, err := r.rdb.Scan(ctx, cursor, "*", 1000).Result()
		if err != nil {
			return totalTargets, fmt.Errorf("failed to read keys: %w", err)
		}

		for _, key := range keys {
			if strings.HasPrefix(key, "lock:") {
				continue
			}

			target, err := r.Get(ctx, key)
			if err != nil {
				slog.Warn("failed to get a value by key from database", "key", err)
				continue
			}

			totalTargets = append(totalTargets, target)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return totalTargets, nil
}
