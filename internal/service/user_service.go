package service

import (
	"context"
	"fmt"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

type iUserRepository interface {
	ExistsByID(ctx context.Context, userID string) (bool, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
}

type UserService struct {
	userRepo iUserRepository
}

func NewUserService(userRepo iUserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	exists, err := s.userRepo.ExistsByID(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("error checking if user %s exists: %w", userID, err)
	}
	if !exists {
		return domain.User{}, fmt.Errorf("user with ID %s: %w", userID, domain.ErrUserNotFound)
	}

	return s.userRepo.SetIsActive(ctx, userID, isActive)
}
