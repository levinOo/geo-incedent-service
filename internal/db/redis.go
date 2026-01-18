package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/levinOo/geo-incedent-service/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(cfg *config.RedisConfig) (*Redis, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
		PoolTimeout:  cfg.RedisPoolTimeout,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
		MaxRetries:   cfg.RedisMaxRetries,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Redis{
		Client: client,
	}, nil
}

func (r *Redis) Close() error {
	r.Client.Close()
	slog.Info("redis connection closed")
	return nil
}
