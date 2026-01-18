package entity

import (
	"time"

	"github.com/google/uuid"
)

type WebhookTask struct {
	ID         uuid.UUID `db:"id"`
	Name       string    `db:"name"`
	UserID     string    `db:"user_id"`
	IncidentID uuid.UUID `db:"incident_id"`
	CreatedAt  time.Time `db:"created_at"`
	RetryCount int       `db:"retry_count"`
}

type WebhookPayload struct {
	Name       string    `json:"name"`
	IncidentID uuid.UUID `json:"incident_id"`
	UserID     string    `json:"user_id"`
	Timestamp  time.Time `json:"timestamp"`
}
