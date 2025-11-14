package repository

import (
	"context"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

type iReviewersPostgres interface {
	AssignReviewersToPR(ctx context.Context, prID string, reviewerIDs []string) error
	GetReviewersByPRID(ctx context.Context, prID string) ([]domain.User, error)
	ReassignReviewer(ctx context.Context, prID, newReviewerID, oldReviewerID string) error
	GetUsersReviewingPR(ctx context.Context, userID string) ([]domain.PullRequest, error)
	IsReviewerAssignedToPR(ctx context.Context, prID, reviewerID string) (bool, error)
}

// ReviewersRepository struct for store interactions related to reviewers
type ReviewersRepository struct {
	postgres iReviewersPostgres
}

func NewReviewersRepository(postgres iReviewersPostgres) *ReviewersRepository {
	return &ReviewersRepository{postgres: postgres}
}

// AssignToPR assigns reviewers with reviewerIDs to a pull request with the given prID
func (r *ReviewersRepository) AssignToPR(ctx context.Context, prID string, reviewerIDs []string) error {
	return r.postgres.AssignReviewersToPR(ctx, prID, reviewerIDs)
}

// GetByPRID retrieves the list of reviewers assigned to a pull request with the given prID
func (r *ReviewersRepository) GetByPRID(ctx context.Context, prID string) ([]domain.User, error) {
	return r.postgres.GetReviewersByPRID(ctx, prID)
}

// Reassign changes the reviewer of a pull request from oldReviewerID to newReviewerID
func (r *ReviewersRepository) Reassign(ctx context.Context, prID, newReviewerID, oldReviewerID string) error {
	return r.postgres.ReassignReviewer(ctx, prID, newReviewerID, oldReviewerID)
}

// GetReviewingPR retrieves the list of pull requests that the user with userID is reviewing
func (r *ReviewersRepository) GetReviewingPR(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	return r.postgres.GetUsersReviewingPR(ctx, userID)
}

// IsReviewerAssignedToPR checks if a reviewer with reviewerID is assigned to a pull request with prID
func (r *ReviewersRepository) IsReviewerAssignedToPR(ctx context.Context, prID, reviewerID string) (bool, error) {
	return r.postgres.IsReviewerAssignedToPR(ctx, prID, reviewerID)
}
