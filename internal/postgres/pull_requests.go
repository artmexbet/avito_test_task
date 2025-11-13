package postgres

import (
	"avito_test_task/internal/domain"
	"avito_test_task/internal/postgres/queries"
	"context"
	"fmt"
)

func (p *Postgres) CreatePullRequest(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := p.queries.WithTx(tx)

	createdPR, err := q.CreatePullRequest(ctx, queries.CreatePullRequestParams{
		ID:       pr.ID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
	})
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error creating pull request: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PullRequest{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return createdPR.ToDomain(), nil
}

func (p *Postgres) GetPullRequestByID(ctx context.Context, prID string) (domain.PullRequest, error) {
	pr, err := p.queries.GetPullRequestByID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error getting pull request by ID: %w", err)
	}

	return pr.ToDomain(), nil
}

func (p *Postgres) MergePullRequest(ctx context.Context, prID string) (domain.PullRequest, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := p.queries.WithTx(tx)

	pr, err := q.MergePullRequest(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error merging pull request: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PullRequest{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return pr.ToDomain(), nil
}
