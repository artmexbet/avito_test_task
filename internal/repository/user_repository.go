package repository

import (
	"github.com/artmexbet/avito_test_task/internal/domain"

	"context"
)

type iUserPostgres interface {
	AddUsers(ctx context.Context, users []domain.User) ([]domain.User, error)
	BatchExistsUserByID(ctx context.Context, users []domain.User) map[domain.User]bool
	ExistsUserByID(ctx context.Context, userID string) (bool, error)
	GetUserByID(ctx context.Context, userID string) (domain.User, error)
	GetUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error)
	SetUserIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	GetActiveUsersByTeamName(ctx context.Context, teamName string) ([]domain.User, error)
}

// UserRepository struct for store interactions related to users
type UserRepository struct {
	postgres iUserPostgres
}

func NewUserRepository(postgres iUserPostgres) *UserRepository {
	return &UserRepository{postgres: postgres}
}

// Add adds new users to the repository
func (r *UserRepository) Add(ctx context.Context, users []domain.User) ([]domain.User, error) {
	return r.postgres.AddUsers(ctx, users)
}

// BatchExistsByID checks the existence of multiple users by their IDs
func (r *UserRepository) BatchExistsByID(ctx context.Context, users []domain.User) map[domain.User]bool {
	return r.postgres.BatchExistsUserByID(ctx, users)
}

// ExistsByID checks if a user exists by their ID
func (r *UserRepository) ExistsByID(ctx context.Context, userID string) (bool, error) {
	return r.postgres.ExistsUserByID(ctx, userID)
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, userID string) (domain.User, error) {
	return r.postgres.GetUserByID(ctx, userID)
}

// GetByTeamName retrieves users by their team name
func (r *UserRepository) GetByTeamName(ctx context.Context, teamName string) ([]domain.User, error) {
	return r.postgres.GetUsersByTeamName(ctx, teamName)
}

// SetIsActive sets the active status of a user
func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	return r.postgres.SetUserIsActive(ctx, userID, isActive)
}

// GetActiveByTeamName retrieves active users by their team name
func (r *UserRepository) GetActiveByTeamName(ctx context.Context, teamName string) ([]domain.User, error) {
	return r.postgres.GetActiveUsersByTeamName(ctx, teamName)
}
