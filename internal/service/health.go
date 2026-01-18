package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/levinOo/geo-incedent-service/internal/db"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/repo/postgres"
	"github.com/redis/go-redis/v9"
)

type HealthService interface {
	Check(ctx context.Context) (*entity.HealthResponse, error)
}

type HealthServiceImpl struct {
	repo  postgres.HealthRepo
	redis *redis.Client
}

func NewHealthService(repo postgres.HealthRepo, redis *db.Redis) *HealthServiceImpl {
	return &HealthServiceImpl{
		repo:  repo,
		redis: redis.Client,
	}
}

func (s *HealthServiceImpl) Check(ctx context.Context) (*entity.HealthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	components := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	checks := map[string]func(context.Context) error{
		"postgres": s.CheckPostgres,
		"redis":    s.CheckRedis,
	}

	hasError := false

	for name, check := range checks {
		wg.Add(1)
		go func(n string, fn func(context.Context) error) {
			defer wg.Done()

			if err := fn(ctx); err != nil {
				mu.Lock()
				components[n] = fmt.Sprintf("error: %s", err)
				hasError = true
				mu.Unlock()
			} else {
				mu.Lock()
				components[n] = "ok"
				mu.Unlock()
			}

		}(name, check)
	}

	wg.Wait()

	status := "ok"
	if hasError {
		status = "error"
	}

	return &entity.HealthResponse{
		Status:     status,
		Components: components,
		Uptime:     time.Now().Format(time.RFC3339),
		Timestamp:  time.Now(),
	}, nil
}

func (s *HealthServiceImpl) CheckPostgres(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *HealthServiceImpl) CheckRedis(ctx context.Context) error {
	return s.redis.Ping(ctx).Err()
}
