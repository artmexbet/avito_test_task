package repository

import (
	"avito_test_task/internal/domain"

	"context"
)

type iPRPostgres interface {
	CreatePullRequest(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error)
	GetPullRequestByID(ctx context.Context, prID string) (domain.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (domain.PullRequest, error)
}

// PRRepository struct for store interactions related to pull requests
type PRRepository struct {
	postgres iPRPostgres
}

func NewPRRepository(postgres iPRPostgres) *PRRepository {
	return &PRRepository{postgres: postgres}
}

// Create adds a new pull request to the repository
func (r *PRRepository) Create(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	return r.postgres.CreatePullRequest(ctx, pr)
}

// GetByID retrieves a pull request by its ID
func (r *PRRepository) GetByID(ctx context.Context, prID string) (domain.PullRequest, error) {
	return r.postgres.GetPullRequestByID(ctx, prID)
}

// Merge merges a pull request by its ID
func (r *PRRepository) Merge(ctx context.Context, prID string) (domain.PullRequest, error) {
	return r.postgres.MergePullRequest(ctx, prID)
}
