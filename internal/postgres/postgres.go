package postgres

import (
	"github.com/artmexbet/avito_test_task/internal/postgres/queries"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate sqlc generate -f ./queries/sqlc.yaml

// Postgres is an abstraction over Postgres database.
type Postgres struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

// New creates a new Postgres instance.
func New(pool *pgxpool.Pool) *Postgres {
	return &Postgres{
		pool:    pool,
		queries: queries.New(pool),
	}
}
