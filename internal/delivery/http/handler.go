package myHttp

import "github.com/levinOo/geo-incedent-service/internal/service"

type Handler struct {
	Incident IncidentHandler
	Location LocationHandler
	Health   HealthHandler
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		Incident: NewIncidentHandler(service),
		Location: NewLocationHandler(service),
		Health:   NewHealthHandler(service),
	}
}
