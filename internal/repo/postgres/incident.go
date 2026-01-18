package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/levinOo/geo-incedent-service/internal/entity"
)

type IncidentRepo interface {
	Create(ctx context.Context, i *entity.Incident) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Incident, error)
	FindAll(ctx context.Context, limit, offset string) ([]entity.Incident, error)
	Update(ctx context.Context, i *entity.Incident) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetStats(ctx context.Context, minutes int) ([]*entity.IncidentStats, error)
	Ping(ctx context.Context) error
}

type IncidentRepoImpl struct {
	pool *pgxpool.Pool
}

func NewIncidentRepo(pool *pgxpool.Pool) IncidentRepo {
	return &IncidentRepoImpl{pool: pool}
}

func (r *IncidentRepoImpl) Create(ctx context.Context, i *entity.Incident) error {
	areaJSON, err := json.Marshal(i.Area)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга area: %w", err)
	}

	query := `
		INSERT INTO incidents (name, description, area, is_active)
		VALUES ($1, $2, ST_GeomFromGeoJSON($3)::geography, $4)
	`

	_, err = r.pool.Exec(ctx, query,
		i.Name, i.Description, string(areaJSON), i.IsActive,
	)
	if err != nil {
		return fmt.Errorf("ошибка создания инцидента: %w", err)
	}

	return nil
}

func (r *IncidentRepoImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Incident, error) {
	var i entity.Incident

	query := `
		SELECT 
			id,
			name,
			description,
			ST_AsGeoJSON(area) AS area_json,
			is_active,
			created_at,
			updated_at
		FROM incidents
		WHERE id = $1
	`

	var areaJSONStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&areaJSONStr,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("инцидент не найден")
		}
		return nil, fmt.Errorf("ошибка поиска инцидента по id %s: %w", id, err)
	}

	if err := json.Unmarshal([]byte(areaJSONStr), &i.Area); err != nil {
		return nil, fmt.Errorf("ошибка размаршалинга area: %w", err)
	}

	return &i, nil
}

func (r *IncidentRepoImpl) FindAll(ctx context.Context, limit, offset string) ([]entity.Incident, error) {
	query := `
		SELECT 
			id,
			name,
			description,
			ST_AsGeoJSON(area) AS area_json,
			is_active,
			created_at,
			updated_at
		FROM incidents
		WHERE is_active = true
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска инцидентов: %w", err)
	}
	defer rows.Close()

	var incidents []entity.Incident

	for rows.Next() {
		var i entity.Incident
		var areaJSONStr string

		err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&areaJSONStr,
			&i.IsActive,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования инцидента: %w", err)
		}

		if err := json.Unmarshal([]byte(areaJSONStr), &i.Area); err != nil {
			return nil, fmt.Errorf("ошибка размаршалинга area: %w", err)
		}

		incidents = append(incidents, i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка rows: %w", err)
	}

	return incidents, nil
}

func (r *IncidentRepoImpl) Update(ctx context.Context, i *entity.Incident) error {
	areaJSON, err := json.Marshal(i.Area)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга area: %w", err)
	}

	query := `
		UPDATE incidents
		SET 
			name = $1,
			description = $2,
			area = ST_GeomFromGeoJSON($3)::geography, 
			updated_at = NOW()
		WHERE id = $4
	`

	_, err = r.pool.Exec(ctx, query,
		i.Name,
		i.Description,
		string(areaJSON),
		i.ID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("инцидент не найден для обновления")
		}
		return fmt.Errorf("ошибка обновления инцидента: %w", err)
	}

	return nil
}

func (r *IncidentRepoImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE incidents
		SET
			is_active = false,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка деактивации инцидента %s: %w", id, err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("инцидент не найден")
	}

	slog.Info("инцидент деактивирован", slog.String("id", id.String()))
	return nil
}

func (r *IncidentRepoImpl) GetStats(ctx context.Context, minutes int) ([]*entity.IncidentStats, error) {
	query := `
        SELECT 
            i.id as incident_id,
            i.name,
            COUNT(DISTINCT lc.user_id) as user_count
        FROM incidents i
        LEFT JOIN location_checks lc ON lc.incident_id = i.id 
            AND lc.is_danger = true
            AND lc.created_at > NOW() - INTERVAL '1 minute' * $1
        WHERE i.is_active = true
        GROUP BY i.id, i.name
        HAVING COUNT(DISTINCT lc.user_id) > 0
        ORDER BY user_count DESC
    `

	rows, err := r.pool.Query(ctx, query, minutes)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса статистики: %w", err)
	}
	defer rows.Close()

	var stats []*entity.IncidentStats
	for rows.Next() {
		var s entity.IncidentStats
		if err := rows.Scan(&s.IncidentID, &s.Name, &s.UserCount); err != nil {
			return nil, fmt.Errorf("ошибка сканирования статистики: %w", err)
		}
		stats = append(stats, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка rows: %w", err)
	}

	return stats, nil
}

func (r *IncidentRepoImpl) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
