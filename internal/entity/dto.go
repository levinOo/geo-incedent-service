package entity

import "time"

type ErrorResponse struct {
	Error   string `json:"error" example:"invalid input"`
	Details string `json:"details,omitempty" example:"email is required"`
}

type HealthResponse struct {
	Status     string            `json:"status"`
	Components map[string]string `json:"components"`
	Uptime     string            `json:"uptime"`
	Timestamp  time.Time         `json:"timestamp"`
}
