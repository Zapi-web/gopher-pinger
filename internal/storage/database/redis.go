package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Zapi-web/gopher-pinger/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisDb struct {
	rdb *redis.Client
}

type DataStruct struct {
	Url             string `redis:"url"`
	LastTimeChecked int    `redis:"last_time_checked"`
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

func (r *RedisDb) Set(ctx context.Context, key string, value DataStruct) error {
	if key == "" || value.Url == "" {
		return domain.ErrInputisEmpty
	}

	err := r.rdb.HSet(ctx, key, value).Err()

	if err != nil {
		return fmt.Errorf("failed to set a value to database %w", err)
	}

	slog.Info("key added to redis", "ULID", key)

	return nil
}

func (r *RedisDb) Get(ctx context.Context, key string) (DataStruct, error) {
	if key == "" {
		return DataStruct{}, domain.ErrInputisEmpty
	}
	var res DataStruct

	err := r.rdb.HGetAll(ctx, key).Scan(&res)

	if err != nil {
		return DataStruct{}, fmt.Errorf("failed to get a value from a database %w", err)
	}

	if res.Url == "" {
		return DataStruct{}, domain.ErrNotFound
	}

	slog.Info("value retrieved from redis", "ULID", key)

	return res, nil
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
