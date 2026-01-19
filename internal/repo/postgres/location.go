package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/levinOo/geo-incedent-service/internal/entity"
)

type LocationRepo interface {
	CheckLocation(ctx context.Context, location entity.UserLocation) ([]*entity.LocationCheckIncident, error)
	SaveLocationCheck(ctx context.Context, location *entity.LocationCheck) error
}

type LocationRepoImpl struct {
	pool *pgxpool.Pool
}

func NewLocationRepo(pool *pgxpool.Pool) LocationRepo {
	return &LocationRepoImpl{pool: pool}
}

func (r *LocationRepoImpl) CheckLocation(ctx context.Context, location entity.UserLocation) ([]*entity.LocationCheckIncident, error) {
	query := `
	SELECT 
		i.id,
		i.name,
		i.description
	FROM incidents i
	WHERE i.is_active = true
	AND ST_Intersects(
		i.area,
		ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
	)
`

	rows, err := r.pool.Query(ctx, query, location.Lon, location.Lat)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки локации: %w", err)
	}
	defer rows.Close()

	var incidents []*entity.LocationCheckIncident
	for rows.Next() {
		var incident entity.LocationCheckIncident
		if err := rows.Scan(&incident.ID, &incident.Name, &incident.Description); err != nil {
			return nil, fmt.Errorf("ошибка сканирования инцидента: %w", err)
		}
		incidents = append(incidents, &incident)
	}
	return incidents, nil
}

func (r *LocationRepoImpl) SaveLocationCheck(ctx context.Context, location *entity.LocationCheck) error {
	query := `
	INSERT INTO location_checks ( user_id, user_location, is_danger, incident_id, created_at)
	VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		location.UserID,
		location.UserLocation.Lon,
		location.UserLocation.Lat,
		location.IsDanger,
		location.IncidentID,
		location.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("ошибка сохранения проверки локации: %w", err)
	}

	return nil
}
