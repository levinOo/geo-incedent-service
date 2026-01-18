package service

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/levinOo/geo-incedent-service/internal/db"
	"github.com/levinOo/geo-incedent-service/internal/service/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHealthService_Check(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(mr *miniredis.Miniredis, repo *mocks.HealthRepo)
		wantStatus string
		wantErrors []string
	}{
		{
			name: "All Systems Operational",
			setup: func(mr *miniredis.Miniredis, repo *mocks.HealthRepo) {
				repo.On("Ping", mock.Anything).Return(nil)
			},
			wantStatus: "ok",
			wantErrors: []string{},
		},
		{
			name: "Postgres Down",
			setup: func(mr *miniredis.Miniredis, repo *mocks.HealthRepo) {
				repo.On("Ping", mock.Anything).Return(errors.New("pg down"))
			},
			wantStatus: "error",
			wantErrors: []string{"postgres"},
		},
		{
			name: "Redis Down",
			setup: func(mr *miniredis.Miniredis, repo *mocks.HealthRepo) {
				repo.On("Ping", mock.Anything).Return(nil)
				mr.Close() // Simulate redis down
			},
			wantStatus: "error",
			wantErrors: []string{"redis"},
		},
		{
			name: "Both Down",
			setup: func(mr *miniredis.Miniredis, repo *mocks.HealthRepo) {
				repo.On("Ping", mock.Anything).Return(errors.New("pg down"))
				mr.Close()
			},
			wantStatus: "error",
			wantErrors: []string{"postgres", "redis"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, err := miniredis.Run()
			require.NoError(t, err)
			t.Cleanup(mr.Close)

			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			dbRedis := &db.Redis{Client: rdb}

			repo := mocks.NewHealthRepo(t)
			tt.setup(mr, repo)

			s := NewHealthService(repo, dbRedis)
			got, err := s.Check(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, got.Status)

			for _, comp := range tt.wantErrors {
				assert.Contains(t, got.Components[comp], "error")
			}

			// Verify healthy components are "ok"
			allComps := []string{"postgres", "redis"}
			for _, c := range allComps {
				isExpectedErr := false
				for _, e := range tt.wantErrors {
					if c == e {
						isExpectedErr = true
						break
					}
				}
				if !isExpectedErr {
					assert.Equal(t, "ok", got.Components[c])
				}
			}
		})
	}
}
