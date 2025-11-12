package postgres

import (
	"avito_test_task/internal/postgres/queries"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate sqlc generate -f ./queries/sqlc.yaml

type Postgres struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

func New(pool *pgxpool.Pool) *Postgres {
	return &Postgres{
		pool:    pool,
		queries: queries.New(pool),
	}
}
