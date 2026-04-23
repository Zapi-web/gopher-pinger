package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Zapi-web/gopher-pinger/internal/domain"
	"github.com/Zapi-web/gopher-pinger/internal/service"
	"github.com/redis/go-redis/v9"
)

type RedisDb struct {
	rdb *redis.Client
}

type dataStruct struct {
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

func (r *RedisDb) Set(ctx context.Context, key string, value service.Target) error {
	if key == "" || value.URL == "" {
		return domain.ErrInputisEmpty
	}

	setValue := dataStruct{
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

func (r *RedisDb) Get(ctx context.Context, key string) (service.Target, error) {
	if key == "" {
		return service.Target{}, domain.ErrInputisEmpty
	}
	var res dataStruct

	err := r.rdb.HGetAll(ctx, key).Scan(&res)

	if err != nil {
		return service.Target{}, fmt.Errorf("failed to get a value from a database %w", err)
	}

	if res.Url == "" {
		return service.Target{}, domain.ErrNotFound
	}

	val := service.Target{
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

func (r *RedisDb) UpdateStatus(ctx context.Context, key string, code int, timestamp string) error {
	err := r.rdb.HSet(ctx, key,
		"last_code", code,
		"last_time_checked", timestamp,
	).Err()

	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}
