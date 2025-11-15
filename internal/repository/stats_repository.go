package repository

import (
	"context"
	"fmt"

	stats_retriever "github.com/artmexbet/avito_test_task/internal/stats-retriever"
)

type iStatsPostgres interface {
	GetUserStats(ctx context.Context) ([]stats_retriever.UsersStats, error)
	GetTeamStats(ctx context.Context) ([]stats_retriever.TeamsStats, error)
	GetAssignmentStats(ctx context.Context) ([]stats_retriever.AssignmentStats, error)
}

type StatsRepository struct {
	postgres iStatsPostgres
}

func NewStatsRepository(postgres iStatsPostgres) *StatsRepository {
	return &StatsRepository{
		postgres: postgres,
	}
}

func (r *StatsRepository) Get(ctx context.Context) ([]stats_retriever.Stats, error) {
	var statsList stats_retriever.Stats
	userStats, err := r.postgres.GetUserStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	statsList.UserStats = userStats

	teamStats, err := r.postgres.GetTeamStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	statsList.TeamStats = teamStats

	assignStats, err := r.postgres.GetAssignmentStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	statsList.AssignStats = assignStats

	return []stats_retriever.Stats{statsList}, nil
}
