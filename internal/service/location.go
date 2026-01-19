package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/levinOo/geo-incedent-service/internal/db"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/queue"
	"github.com/levinOo/geo-incedent-service/internal/repo/postgres"
	"github.com/levinOo/geo-incedent-service/pkg/validator"
)

const (
	cacheKey       = "incidents:active"
	contextTimeout = 2 * time.Second
)

type LocationService interface {
	CheckLocation(ctx context.Context, req *entity.CheckLocationRequest) (*entity.CheckLocationResponse, error)
}

type LocationServiceImpl struct {
	repo         postgres.LocationRepo
	incidentRepo postgres.IncidentRepo
	queue        queue.Queue
	redis        *db.Redis
}

func NewLocationService(repo postgres.LocationRepo, incidentRepo postgres.IncidentRepo, redis *db.Redis) LocationService {
	return &LocationServiceImpl{
		repo:         repo,
		incidentRepo: incidentRepo,
		queue:        *queue.NewQueue(redis.Client),
		redis:        redis,
	}
}

func (s *LocationServiceImpl) CheckLocation(ctx context.Context, req *entity.CheckLocationRequest) (*entity.CheckLocationResponse, error) {
	if err := validator.ValidateLocation(req.UserLocation); err != nil {
		slog.Error("ошибка валидации локации", "error", err)
		return nil, fmt.Errorf("ошибка валидации локации: %w", err)
	}

	var incidents []entity.Incident

	cachedData, err := s.redis.Client.Get(ctx, cacheKey).Bytes()
	if err == nil {
		if err := json.Unmarshal(cachedData, &incidents); err != nil {
			slog.Error("ошибка десериализации (unmarshal) инцидентов из кэша", "error", err)
		}
	}

	if len(incidents) == 0 {
		incidents, err = s.incidentRepo.FindAll(ctx, 1000, 0)
		if err != nil {
			slog.Error("не удалось получить активные инциденты", "error", err)
			return nil, fmt.Errorf("ошибка получения активных инцидентов: %w", err)
		}

		if data, err := json.Marshal(incidents); err == nil {
			s.redis.Client.Set(ctx, cacheKey, data, 1*time.Minute)
		}
	}

	var matchedIncidents []*entity.LocationCheckIncident
	slog.Debug("Проверка инцидентов", "count", len(incidents))
	for _, inc := range incidents {
		slog.Debug("Проверка инцидента", "name", inc.Name, "area", inc.Area)
		if inc.Area.Contains(req.UserLocation.Lat, req.UserLocation.Lon) {
			slog.Info("Инцидент найден", "name", inc.Name)
			matchedIncidents = append(matchedIncidents, &entity.LocationCheckIncident{
				ID:          inc.ID,
				Name:        inc.Name,
				Description: inc.Description,
			})
		}
	}

	isDanger := len(matchedIncidents) > 0
	var incidentID *uuid.UUID
	if isDanger {
		id := matchedIncidents[0].ID
		incidentID = &id
	}

	check := &entity.LocationCheck{
		UserID:       req.UserID,
		UserLocation: req.UserLocation,
		IsDanger:     isDanger,
		IncidentID:   incidentID,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.SaveLocationCheck(ctx, check); err != nil {
		slog.Error("не удалось сохранить проверку локации", "error", err)
		return nil, fmt.Errorf("ошибка сохранения проверки локации: %w", err)
	}

	if isDanger && len(matchedIncidents) > 0 {
		task := &entity.WebhookTask{
			ID:         uuid.New(),
			Name:       matchedIncidents[0].Name,
			UserID:     req.UserID,
			IncidentID: *incidentID,
			CreatedAt:  time.Now(),
		}

		enqueueCtx, cancel := context.WithTimeout(ctx, contextTimeout)
		defer cancel()

		if err := s.queue.Enqueue(enqueueCtx, task); err != nil {
			slog.Error("ошибка добавления вебхука в очередь", "error", err)
		}
	}

	return &entity.CheckLocationResponse{
		IsDanger:  isDanger,
		Incidents: matchedIncidents,
	}, nil
}
