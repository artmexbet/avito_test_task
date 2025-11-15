package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/artmexbet/avito_test_task/internal/domain"
	"github.com/artmexbet/avito_test_task/internal/postgres/queries"
)

func (p *Postgres) AssignReviewersToPR(ctx context.Context, prID string, reviewerIDs []string) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := p.queries.WithTx(tx)

	params := make([]queries.AssignReviewerToPullRequestParams, len(reviewerIDs))
	for i, reviewerID := range reviewerIDs {
		params[i] = queries.AssignReviewerToPullRequestParams{
			PullRequestID: prID,
			ReviewerID:    reviewerID,
		}
	}
	br := q.AssignReviewerToPullRequest(ctx, params)
	defer br.Close() //nolint:errcheck

	errs := make([]error, 0, len(reviewerIDs))
	br.QueryRow(func(_ int, r queries.PullRequestsReviewer, err error) {
		if err != nil {
			errs = append(errs, err)
		}
	})
	err = errors.Join(errs...)
	if err != nil {
		return fmt.Errorf("error assigning reviewers to PR: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	return nil
}

func (p *Postgres) GetReviewersByPRID(ctx context.Context, prID string) ([]domain.User, error) {
	users, err := p.queries.GetReviewersByPullRequestID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("error getting reviewers by PR ID: %w", err)
	}

	var reviewers []domain.User
	for _, u := range users {
		reviewers = append(reviewers, u.ToDomain())
	}
	return reviewers, nil
}

func (p *Postgres) ReassignReviewer(ctx context.Context, prID, newReviewerID, oldReviewerID string) error {
	return p.queries.ReassignReviewerForPullRequest(ctx, queries.ReassignReviewerForPullRequestParams{
		PullRequestID: prID,
		ReviewerID:    newReviewerID,
		ReviewerID_2:  oldReviewerID,
	})
}

func (p *Postgres) GetUsersReviewingPR(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	prs, err := p.queries.GetUsersReviewingPullRequest(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting PRs being reviewed by user: %w", err)
	}

	var pullRequests []domain.PullRequest
	for _, pr := range prs {
		pullRequests = append(pullRequests, pr.ToDomain())
	}
	return pullRequests, nil
}

func (p *Postgres) IsReviewerAssignedToPR(ctx context.Context, prID, reviewerID string) (bool, error) {
	assigned, err := p.queries.IsUserReviewerForPullRequest(ctx, queries.IsUserReviewerForPullRequestParams{
		PullRequestID: prID,
		ReviewerID:    reviewerID,
	})
	if err != nil {
		return false, err
	}

	return assigned, nil
}
