package postgres

import (
	"github.com/artmexbet/avito_test_task/internal/domain"

	"context"
	"fmt"
)

func (p *Postgres) GetTeamByName(ctx context.Context, teamName string) (domain.Team, error) {
	team, err := p.queries.GetTeamByName(ctx, teamName)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to get team by name: %w", err)
	}
	return team.ToDomain(), nil
}

func (p *Postgres) AddTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := p.queries.WithTx(tx)

	dbTeam, err := qtx.AddTeam(ctx, team.Name)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to add team: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Team{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dbTeam.ToDomain(), nil
}

func (p *Postgres) ExistsTeamByName(ctx context.Context, teamName string) (bool, error) {
	exists, err := p.queries.ExistsTeamByName(ctx, teamName)
	if err != nil {
		return false, fmt.Errorf("failed to check if team exists by name: %w", err)
	}
	return exists, nil
}
