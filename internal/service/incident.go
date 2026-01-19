package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/levinOo/geo-incedent-service/config"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/repo/postgres"
	"github.com/levinOo/geo-incedent-service/pkg/validator"
)

type IncidentService interface {
	Create(ctx context.Context, req *entity.CreateIncidentRequest) (*entity.IncidentResponse, error)
	FindByID(ctx context.Context, id string) (*entity.GetIncidentResponse, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entity.GetIncidentResponse, error)
	Update(ctx context.Context, req *entity.UpdateIncidentRequest, id string) (*entity.IncidentResponse, error)
	Delete(ctx context.Context, id string) (*entity.IncidentResponse, error)
	GetStats(ctx context.Context) (*entity.StatsResponse, error)
}

type IncidentServiceImpl struct {
	repo postgres.IncidentRepo
	cfg  *config.Config
}

func NewIncidentService(repo postgres.IncidentRepo, cfg *config.Config) IncidentService {
	return &IncidentServiceImpl{repo: repo, cfg: cfg}
}

func (s *IncidentServiceImpl) Create(ctx context.Context, req *entity.CreateIncidentRequest) (*entity.IncidentResponse, error) {
	if err := validator.ValidatePolygon(req.Area); err != nil {
		slog.Error("ошибка валидации полигона", "error", err.Error())
		return nil, fmt.Errorf("ошибка валидации полигона: %w", err)
	}

	incident := &entity.Incident{
		Name:        req.Name,
		Description: req.Description,
		Area:        req.Area,
		IsActive:    true,
	}

	err := s.repo.Create(ctx, incident)
	if err != nil {
		slog.Error("не удалось создать инцидент", "error", err.Error())
		return nil, fmt.Errorf("не удалось создать инцидент: %w", err)
	}

	return &entity.IncidentResponse{
		Status: "успешно создан",
	}, nil
}

func (s *IncidentServiceImpl) FindByID(ctx context.Context, id string) (*entity.GetIncidentResponse, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		slog.Error("ошибка парсинга uuid", "error", err.Error())
		return nil, fmt.Errorf("ошибка парсинга uuid: %w", err)
	}

	incident, err := s.repo.FindByID(ctx, uuid)
	if err != nil {
		slog.Error("не удалось найти инцидент", "error", err.Error())
		return nil, fmt.Errorf("не удалось найти инцидент: %w", err)
	}

	return &entity.GetIncidentResponse{
		ID:          incident.ID.String(),
		Name:        incident.Name,
		Description: incident.Description,
		Area:        incident.Area,
		IsActive:    incident.IsActive,
		CreatedAt:   incident.CreatedAt,
		UpdatedAt:   incident.UpdatedAt,
	}, nil
}

func (s *IncidentServiceImpl) FindAll(ctx context.Context, limit, offset int) ([]*entity.GetIncidentResponse, error) {
	incidents, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		slog.Error("не удалось найти инциденты", "error", err.Error())
		return nil, fmt.Errorf("не удалось найти инциденты: %w", err)
	}

	var incidentResponses []*entity.GetIncidentResponse
	for _, incident := range incidents {
		incidentResponses = append(incidentResponses, &entity.GetIncidentResponse{
			ID:          incident.ID.String(),
			Name:        incident.Name,
			Description: incident.Description,
			Area:        incident.Area,
			IsActive:    incident.IsActive,
			CreatedAt:   incident.CreatedAt,
			UpdatedAt:   incident.UpdatedAt,
		})
	}

	return incidentResponses, nil
}

func (s *IncidentServiceImpl) Update(ctx context.Context, req *entity.UpdateIncidentRequest, id string) (*entity.IncidentResponse, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		slog.Error("ошибка парсинга uuid", "error", err)
		return nil, fmt.Errorf("ошибка парсинга uuid: %w", err)
	}

	if req.Name == nil && req.Description == nil && req.Area == nil {
		slog.Error("не указаны поля для обновления")
		return nil, fmt.Errorf("не указаны поля для обновления")
	}

	if req.Area != nil {
		if err := validator.ValidatePolygon(*req.Area); err != nil {
			slog.Error("некорректный полигон", "error", err)
			return nil, fmt.Errorf("некорректный полигон: %w", err)
		}
	}

	currentIncident, err := s.repo.FindByID(ctx, uuid)
	if err != nil {
		slog.Error("не удалось найти инцидент", "error", err)
		return nil, fmt.Errorf("не удалось найти инцидент: %w", err)
	}

	if req.Name != nil {
		currentIncident.Name = *req.Name
	}

	if req.Description != nil {
		currentIncident.Description = *req.Description
	}

	if req.Area != nil {
		currentIncident.Area = *req.Area
	}

	if err := s.repo.Update(ctx, currentIncident); err != nil {
		slog.Error("не удалось обновить инцидент", "error", err)
		return nil, fmt.Errorf("не удалось обновить инцидент: %w", err)
	}

	return &entity.IncidentResponse{
		Status: "успешно обновлен",
	}, nil
}

func (s *IncidentServiceImpl) Delete(ctx context.Context, id string) (*entity.IncidentResponse, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		slog.Error("ошибка парсинга uuid", "error", err)
		return nil, fmt.Errorf("ошибка парсинга uuid: %w", err)
	}

	if err := s.repo.Delete(ctx, uuid); err != nil {
		slog.Error("не удалось удалить инцидент", "error", err)
		return nil, fmt.Errorf("не удалось удалить инцидент: %w", err)
	}

	return &entity.IncidentResponse{
		Status: "успешно удален",
	}, nil
}

func (s *IncidentServiceImpl) GetStats(ctx context.Context) (*entity.StatsResponse, error) {
	stats, err := s.repo.GetStats(ctx, s.cfg.HTTPServer.StatsWindowMinutes)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить статистику: %w", err)
	}

	return &entity.StatsResponse{
		Stats:         stats,
		WindowMinutes: s.cfg.HTTPServer.StatsWindowMinutes,
	}, nil
}
