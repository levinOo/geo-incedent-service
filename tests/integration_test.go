package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/levinOo/geo-incedent-service/config"
	"github.com/levinOo/geo-incedent-service/internal/db"
	myHttp "github.com/levinOo/geo-incedent-service/internal/delivery/http"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/repo"
	"github.com/levinOo/geo-incedent-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_FullFlow(t *testing.T) {
	// Root Logger for Debugging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := &config.Config{
		Postgres: config.PostgresConfig{
			PostgresURL: "postgres://postgres:postgres@localhost:5434/geo_incidents?sslmode=disable",
		},
		Redis: config.RedisConfig{
			RedisAddr: "localhost:6379",
		},
		HTTPServer: config.HTTPServerConfig{
			HTTPServerPort:     "8080",
			StatsWindowMinutes: 60,
			APIKey:             "test-api-key",
		},
		Worker: config.Worker{
			WebhookURL: "http://localhost:9090",
			MaxRetries: 5,
		},
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.Postgres.PostgresURL)
	if err != nil {
		t.Skipf("Postgres connection failed: %v", err)
	}
	defer pool.Close()

	redisClient, err := db.NewRedis(&cfg.Redis)
	if err != nil {
		t.Skipf("Redis connection failed: %v", err)
	}
	defer redisClient.Close()

	// Clean up state
	pool.Exec(ctx, "TRUNCATE incidents, location_checks CASCADE")
	redisClient.Client.FlushAll(ctx)

	repository := repo.NewRepo(pool)
	svc := service.NewService(repository, cfg, redisClient)
	router := myHttp.NewRouter(&cfg.HTTPServer, svc)

	// Step A: Create Incident
	incidentName := fmt.Sprintf("Test Incident %d", time.Now().UnixNano())
	createReq := entity.CreateIncidentRequest{
		Name:        incidentName,
		Description: "Integration test incident",
		Area: entity.GeoJsonPolygon{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{37.0, 55.0},
					{38.0, 55.0},
					{38.0, 56.0},
					{37.0, 56.0},
					{37.0, 55.0},
				},
			},
		},
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/incidents", bytes.NewReader(body))
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, "Incident creation failed: %s", w.Body.String())

	// Step B: Check Location (Should be Danger)
	checkReq := entity.CheckLocationRequest{
		UserID: uuid.NewString(),
		UserLocation: entity.UserLocation{
			Lat: 55.5,
			Lon: 37.5,
		},
	}
	body, _ = json.Marshal(checkReq)
	req = httptest.NewRequest("POST", "/api/v1/location/check", bytes.NewReader(body))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "Location check failed: %s", w.Body.String())

	var checkResp entity.CheckLocationResponse
	err = json.Unmarshal(w.Body.Bytes(), &checkResp)
	require.NoError(t, err)
	assert.True(t, checkResp.IsDanger, "User should be in danger zone")
	assert.NotEmpty(t, checkResp.Incidents, "Matched incidents should not be empty")

	// Step C: Check Stats
	time.Sleep(200 * time.Millisecond)
	req = httptest.NewRequest("GET", "/api/v1/incidents/stats", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "Stats check failed: %s", w.Body.String())

	var statsResp entity.StatsResponse
	err = json.Unmarshal(w.Body.Bytes(), &statsResp)
	require.NoError(t, err)
	assert.NotEmpty(t, statsResp.Stats, "Stats should not be empty")

	// Step D: Check Health
	req = httptest.NewRequest("GET", "/api/v1/system/health", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "Health check failed: %s", w.Body.String())
}
