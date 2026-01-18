package service

import (
	"github.com/levinOo/geo-incedent-service/config"
	"github.com/levinOo/geo-incedent-service/internal/db"
	"github.com/levinOo/geo-incedent-service/internal/repo"
)

type Service struct {
	Incident IncidentService
	Location LocationService
	Health   HealthService
}

func NewService(repo *repo.Repo, cfg *config.Config, redis *db.Redis) *Service {
	return &Service{
		Incident: NewIncidentService(repo.IncidentRepo, cfg),
		Location: NewLocationService(repo.LocationRepo, repo.IncidentRepo, redis),
		Health:   NewHealthService(repo.HealthRepo, redis),
	}
}
