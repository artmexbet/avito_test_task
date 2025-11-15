package postgres

import (
	"context"
	"fmt"

	stats_retriever "github.com/artmexbet/avito_test_task/internal/stats-retriever"
)

func (p *Postgres) GetUserStats(ctx context.Context) ([]stats_retriever.UsersStats, error) {
	res, err := p.queries.GetUsersCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetUserStats: %w", err)
	}

	var userStats []stats_retriever.UsersStats
	for _, r := range res {
		userStats = append(userStats, stats_retriever.UsersStats{
			IsActive: r.IsActive,
			Total:    int(r.UserCount),
		})
	}
	return userStats, nil
}

func (p *Postgres) GetTeamStats(ctx context.Context) ([]stats_retriever.TeamsStats, error) {
	res, err := p.queries.GetTeamsCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetTeamStats: %w", err)
	}

	var teamStats []stats_retriever.TeamsStats
	for _, r := range res {
		teamStats = append(teamStats, stats_retriever.TeamsStats{
			TeamName: r.Name,
			TotalPRs: int(r.PrCount),
		})
	}
	return teamStats, nil
}

func (p *Postgres) GetAssignmentStats(ctx context.Context) ([]stats_retriever.AssignmentStats, error) {
	res, err := p.queries.GetAssignmentStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAssignmentStats: %w", err)
	}

	var assignStats []stats_retriever.AssignmentStats
	for _, r := range res {
		assignStats = append(assignStats, stats_retriever.AssignmentStats{
			ReviewerID: r.ReviewerID,
			IsActive:   r.IsActive,
			PRCount:    int(r.AssignedPullRequests),
		})
	}
	return assignStats, nil
}
