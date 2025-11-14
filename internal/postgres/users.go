package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/artmexbet/avito_test_task/internal/domain"
	"github.com/artmexbet/avito_test_task/internal/postgres/queries"
)

func (p *Postgres) AddUsers(ctx context.Context, users []domain.User) ([]domain.User, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck  // safe to call even after commit
	q := p.queries.WithTx(tx)

	params := make([]queries.AddUsersParams, len(users))
	for i, user := range users {
		params[i] = queries.AddUsersParams{
			ID:       user.ID,
			Username: user.Username,
			TeamName: user.TeamName,
		}
	}
	br := q.AddUsers(ctx, params)
	defer br.Close() //nolint:errcheck
	addedUsers := make([]domain.User, len(users))
	errs := make([]error, 0, len(users))
	br.QueryRow(func(i int, user queries.User, err error) {
		addedUsers[i] = user.ToDomain()
		if err != nil {
			errs = append(errs, err)
		}
	})
	err = errors.Join(errs...)
	if err != nil {
		return nil, fmt.Errorf("error adding users: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}
	return addedUsers, nil
}

func (p *Postgres) BatchExistsUserByID(ctx context.Context, users []domain.User) map[domain.User]bool {
	params := make([]string, len(users))
	for i, user := range users {
		params[i] = user.ID
	}
	br := p.queries.BatchExistsUserByID(ctx, params)
	defer br.Close() //nolint:errcheck
	existsMap := make(map[domain.User]bool, len(users))
	br.QueryRow(func(i int, exists bool, _ error) {
		existsMap[users[i]] = exists
	})
	return existsMap
}

func (p *Postgres) ExistsUserByID(ctx context.Context, userID string) (bool, error) {
	return p.queries.ExistsUserByID(ctx, userID)
}

func (p *Postgres) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	user, err := p.queries.GetUserByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	return user.ToDomain(), nil
}

func (p *Postgres) GetUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error) {
	users, err := p.queries.GetUsersByTeamName(ctx, teamName)
	if err != nil {
		return nil, err
	}
	domainUsers := make([]domain.User, len(users))
	for i, user := range users {
		domainUsers[i] = user.ToDomain()
	}
	return domainUsers, nil
}

func (p *Postgres) SetUserIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck  // safe to call even after commit
	q := p.queries.WithTx(tx)

	user, err := q.SetUserIsActiveByID(ctx, queries.SetUserIsActiveByIDParams{
		ID:       userID,
		IsActive: isActive,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("error setting user %s is_active: %w", userID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("error committing transaction: %w", err)
	}
	return user.ToDomain(), nil
}

func (p *Postgres) GetActiveUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error) {
	users, err := p.queries.GetActiveUsersByTeamName(ctx, teamName)
	if err != nil {
		return nil, err
	}
	domainUsers := make([]domain.User, len(users))
	for i, user := range users {
		domainUsers[i] = user.ToDomain()
	}
	return domainUsers, nil
}
