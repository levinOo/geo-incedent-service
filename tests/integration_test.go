package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	// Skip if short mode (optional, but good practice)
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// 1. Setup Configuration
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
			WebhookURL:         "http://localhost:9090",
			APIKey:             "test-api-key",
		},
	}

	// 2. Connect to Dependencies
	pgURL := cfg.Postgres.PostgresURL

	pool, err := pgxpool.New(context.Background(), pgURL)
	if err != nil {
		t.Logf("Skipping integration test: Postgres connection failed: %v", err)
		t.SkipNow()
	}
	defer pool.Close()

	redisClient, err := db.NewRedis(&cfg.Redis)
	if err != nil {
		t.Logf("Skipping integration test: Redis connection failed: %v", err)
		t.SkipNow()
	}
	defer redisClient.Close()

	// 3. Initialize App Layers
	repository := repo.NewRepo(pool)
	// Truncate/Clean DB? Or just use unique data.
	// Let's use unique data to avoid conflicts.

	svc := service.NewService(repository, cfg, redisClient)
	router := myHttp.NewRouter(&cfg.HTTPServer, svc)

	// 4. Test Scenario
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
	req.Header.Set("Authorization", "test-api-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var createResp entity.IncidentResponse
	err = json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)
	assert.Equal(t, "successfully created", createResp.Status)

	// Step B: Check Location (Should be Danger)
	checkReq := entity.CheckLocationRequest{
		UserID: "user-123",
		UserLocation: entity.UserLocation{
			Lat: 55.5,
			Lon: 37.5,
		},
	}
	body, _ = json.Marshal(checkReq)
	req = httptest.NewRequest("POST", "/api/v1/location", bytes.NewReader(body))
	// No Auth needed for location? Check router.go.
	// Usually location is public or authenticated?
	// Router says: `api.POST("/location", ...)` inside `api := router.Group("/api/v1")`.
	// Middleware `AuthMiddleware` is applied to `api`?
	// `router.Use(AuthMiddleware(cfg.APIKey))` -> Yes, global auth.
	req.Header.Set("Authorization", "test-api-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var checkResp entity.CheckLocationResponse
	err = json.Unmarshal(w.Body.Bytes(), &checkResp)
	require.NoError(t, err)
	assert.True(t, checkResp.IsDanger, "User should be in danger zone")
	// Verify incidents list?
	assert.NotEmpty(t, checkResp.Incidents)

	// Step C: Check Stats
	req = httptest.NewRequest("GET", "/api/v1/incidents/stats", nil)
	req.Header.Set("Authorization", "test-api-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var statsResp entity.StatsResponse
	err = json.Unmarshal(w.Body.Bytes(), &statsResp)
	require.NoError(t, err)
	assert.NotEmpty(t, statsResp.Stats)
	// Found our incident?
	found := false
	for _, s := range statsResp.Stats {
		if s.Name == incidentName {
			found = true
			assert.GreaterOrEqual(t, s.UserCount, 1)
			break
		}
	}
	assert.True(t, found, "Stats should contain the test incident")

}
