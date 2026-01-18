package entity

import "github.com/google/uuid"

type IncidentStats struct {
	IncidentID uuid.UUID `json:"incident_id"`
	Name       string    `json:"name"`
	UserCount  int       `json:"user_count"`
}

type StatsResponse struct {
	Stats         []*IncidentStats `json:"stats"`
	WindowMinutes int              `json:"window_minutes" example:"60"`
}
