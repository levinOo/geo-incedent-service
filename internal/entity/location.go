package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type LocationCheck struct {
	ID           uuid.UUID    `db:"id"`
	UserID       string       `db:"user_id"`
	UserLocation UserLocation `db:"-"`
	IsDanger     bool         `db:"is_danger"`
	IncidentID   *uuid.UUID   `db:"incident_id"`
	CreatedAt    time.Time    `db:"created_at"`
}

type LocationCheckIncident struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

type CheckLocationRequest struct {
	UserID       string       `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserLocation UserLocation `json:"user_location" binding:"required"`
}

type CheckLocationResponse struct {
	IsDanger  bool                     `json:"is_danger" example:"true"`
	Incidents []*LocationCheckIncident `json:"incidents,omitempty"`
}
