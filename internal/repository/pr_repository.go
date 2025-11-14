package repository

import (
	"fmt"

	"github.com/artmexbet/avito_test_task/internal/domain"

	"context"
)

type iPRPostgres interface {
	CreatePullRequest(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error)
	GetPullRequestByID(ctx context.Context, prID string) (domain.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (domain.PullRequest, error)
	ExistsPullRequest(ctx context.Context, prID string) (bool, error)
	GetReviewersByPRID(ctx context.Context, prID string) ([]domain.User, error)
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
	pr, err := r.postgres.GetPullRequestByID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error getting pull request by ID: %w", err)
	}
	pr.Reviewers, err = r.postgres.GetReviewersByPRID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error getting pull request reviewers: %w", err)
	}
	return pr, nil
}

// Merge merges a pull request by its ID
func (r *PRRepository) Merge(ctx context.Context, prID string) (domain.PullRequest, error) {
	return r.postgres.MergePullRequest(ctx, prID)
}

// Exists checks if a pull request with the given ID exists
func (r *PRRepository) Exists(ctx context.Context, prID string) (bool, error) {
	return r.postgres.ExistsPullRequest(ctx, prID)
}
