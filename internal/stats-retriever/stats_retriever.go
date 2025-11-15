package stats_retriever

import (
	"context"
)

// Скрываю всё абстракциями, в README опишу, почему сделал так.

type iStatsRepository interface {
	Get(ctx context.Context) ([]Stats, error)
}

type StatsRetriever struct {
	repo iStatsRepository
}

func NewStatsRetriever(repo iStatsRepository) *StatsRetriever {
	return &StatsRetriever{
		repo: repo,
	}
}

func (sr *StatsRetriever) RetrieveStats(ctx context.Context) ([]Stats, error) {
	return sr.repo.Get(ctx)
}
