package entity

import (
	"time"

	"github.com/google/uuid"
)

type Incident struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description,omitempty" db:"description"`
	Area        GeoJsonPolygon `json:"area" db:"area"`
	IsActive    bool           `json:"is_active" db:"is_active"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

type CreateIncidentRequest struct {
	Name        string         `json:"name" binding:"required,min=1,max=255" example:"Наводнение"`
	Description string         `json:"description" binding:"omitempty,max=1000" example:"Описание наводнения"`
	Area        GeoJsonPolygon `json:"area" binding:"required"`
}

type UpdateIncidentRequest struct {
	Name        *string         `json:"name" binding:"omitempty,min=1,max=255" example:"Наводнение"`
	Description *string         `json:"description" binding:"omitempty,max=1000" example:"Описание наводнения"`
	Area        *GeoJsonPolygon `json:"area" binding:"omitempty"`
}

type IncidentResponse struct {
	Status string `json:"status" example:"успешно создано"`
	Error  string `json:"error,omitempty" example:"ошибка ввода: неверная область"`
}

type GetIncidentResponse struct {
	ID          string         `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string         `json:"name" example:"Наводнение"`
	Description string         `json:"description" example:"Описание наводнения"`
	Area        GeoJsonPolygon `json:"area"`
	IsActive    bool           `json:"is_active" example:"true"`
	CreatedAt   time.Time      `json:"created_at" example:"2026-01-18T18:30:00Z"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2026-01-18T18:30:00Z"`
}

type GetIncidentsResponse struct {
	Incidents []GetIncidentResponse `json:"incidents"`
	Total     int                   `json:"total" example:"10"`
}
