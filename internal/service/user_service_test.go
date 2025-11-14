package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

// UserServiceTestSuite определяет test suite для UserService
type UserServiceTestSuite struct {
	suite.Suite
	ctx context.Context
}

// SetupTest выполняется перед каждым тестом
func (s *UserServiceTestSuite) SetupTest() {
	s.ctx = context.Background()
}

// TestUserService_SetIsActive проверяет метод SetIsActive
func (s *UserServiceTestSuite) TestSetIsActive() {
	tests := []struct {
		name        string
		userID      string
		isActive    bool
		arrangeFunc func(ctx context.Context, mockRepo *mockiUserRepository)
		wantErr     bool
		wantErrIs   error
		checkResult func(result domain.User)
	}{
		{
			name:     "success - activate user",
			userID:   "user-123",
			isActive: true,
			arrangeFunc: func(ctx context.Context, mockRepo *mockiUserRepository) {
				mockRepo.EXPECT().
					ExistsByID(ctx, "user-123").
					Return(true, nil).Once()

				mockRepo.EXPECT().
					SetIsActive(ctx, "user-123", true).
					Return(domain.User{
						ID:       "user-123",
						Username: "testuser",
						IsActive: true,
					}, nil).Once()
			},
			wantErr: false,
			checkResult: func(result domain.User) {
				s.Equal("user-123", result.ID)
				s.True(result.IsActive)
			},
		},
		{
			name:     "success - deactivate user",
			userID:   "user-123",
			isActive: false,
			arrangeFunc: func(ctx context.Context, mockRepo *mockiUserRepository) {
				mockRepo.EXPECT().
					ExistsByID(ctx, "user-123").
					Return(true, nil).Once()
				mockRepo.EXPECT().
					SetIsActive(ctx, "user-123", false).
					Return(domain.User{
						ID:       "user-123",
						Username: "testuser",
						IsActive: false,
					}, nil).Once()
			},
			wantErr: false,
			checkResult: func(result domain.User) {
				s.Equal("user-123", result.ID)
				s.False(result.IsActive)
			},
		},
		{
			name:     "user not found",
			userID:   "non-existent-user",
			isActive: false,
			arrangeFunc: func(ctx context.Context, mockRepo *mockiUserRepository) {
				mockRepo.EXPECT().
					ExistsByID(ctx, "non-existent-user").
					Return(false, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrUserNotFound,
		},
		{
			name:     "exists check error",
			userID:   "user-123",
			isActive: true,
			arrangeFunc: func(ctx context.Context, mockRepo *mockiUserRepository) {
				mockRepo.EXPECT().
					ExistsByID(ctx, "user-123").
					Return(false, errors.New("database connection error")).Once()
			},
			wantErr: true,
		},
		{
			name:     "set active error",
			userID:   "user-123",
			isActive: false,
			arrangeFunc: func(ctx context.Context, mockRepo *mockiUserRepository) {
				mockRepo.EXPECT().
					ExistsByID(ctx, "user-123").
					Return(true, nil).Once()
				mockRepo.EXPECT().
					SetIsActive(ctx, "user-123", false).
					Return(domain.User{}, errors.New("update error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			mockRepo := newMockiUserRepository(s.T())
			service := NewUserService(mockRepo)

			tt.arrangeFunc(s.ctx, mockRepo)

			// Act
			result, err := service.SetIsActive(s.ctx, tt.userID, tt.isActive)

			// Assert
			if tt.wantErr {
				s.Error(err)
				if tt.wantErrIs != nil {
					s.ErrorIs(err, tt.wantErrIs)
				}
				s.Equal(domain.User{}, result)
			} else {
				s.NoError(err)
				if tt.checkResult != nil {
					tt.checkResult(result)
				}
			}
		})
	}
}

// TestUserServiceSuite запускает test suite
func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
