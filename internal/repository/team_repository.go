package repository

import (
	"context"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

type iTeamPostgres interface {
	GetTeamByName(ctx context.Context, teamName string) (domain.Team, error)
	AddTeam(ctx context.Context, team domain.Team) (domain.Team, error)
	ExistsTeamByName(ctx context.Context, teamName string) (bool, error)
}

type TeamRepository struct {
	postgres iTeamPostgres
}

func NewTeamRepository(postgres iTeamPostgres) *TeamRepository {
	return &TeamRepository{postgres: postgres}
}

func (r *TeamRepository) Get(ctx context.Context, teamName string) (domain.Team, error) {
	return r.postgres.GetTeamByName(ctx, teamName)
}

func (r *TeamRepository) Add(ctx context.Context, team domain.Team) (domain.Team, error) {
	return r.postgres.AddTeam(ctx, team)
}

func (r *TeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	return r.postgres.ExistsTeamByName(ctx, teamName)
}
