package repo

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/levinOo/geo-incedent-service/internal/repo/postgres"
)

type Repo struct {
	IncidentRepo postgres.IncidentRepo
	LocationRepo postgres.LocationRepo
	HealthRepo   postgres.HealthRepo
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		IncidentRepo: postgres.NewIncidentRepo(pool),
		LocationRepo: postgres.NewLocationRepo(pool),
		HealthRepo:   postgres.NewHealthRepoImpl(pool),
	}
}
