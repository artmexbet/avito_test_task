package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

// TeamServiceTestSuite определяет test suite для TeamService
type TeamServiceTestSuite struct {
	suite.Suite
	ctx context.Context
}

// SetupTest выполняется перед каждым тестом
func (s *TeamServiceTestSuite) SetupTest() {
	s.ctx = context.Background()
}

// TestAdd проверяет метод Add
func (s *TeamServiceTestSuite) TestAdd() {
	tests := []struct {
		name        string
		team        domain.Team
		arrangeFunc func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository)
		wantErr     bool
		wantErrIs   error
		checkResult func(result domain.Team)
	}{
		{
			name: "success - all users exist",
			team: domain.Team{
				Name: "backend-team",
				Members: []domain.User{
					{ID: "user-1", Username: "alice"},
					{ID: "user-2", Username: "bob"},
				},
			},
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Exists(ctx, "backend-team").Return(false, nil).Once()
				mockTeamRepo.EXPECT().Add(ctx, domain.Team{
					Name: "backend-team",
					Members: []domain.User{
						{ID: "user-1", Username: "alice"},
						{ID: "user-2", Username: "bob"},
					},
				}).Return(domain.Team{Name: "backend-team"}, nil).Once()
				mockUserRepo.EXPECT().BatchExistsByID(ctx, []domain.User{
					{ID: "user-1", Username: "alice"},
					{ID: "user-2", Username: "bob"},
				}).Return(map[domain.User]bool{
					{ID: "user-1", Username: "alice"}: true,
					{ID: "user-2", Username: "bob"}:   true,
				}).Once()
			},
			wantErr: false,
			checkResult: func(result domain.Team) {
				s.Equal("backend-team", result.Name)
			},
		},
		{
			name: "success - with new users",
			team: domain.Team{
				Name: "new-team",
				Members: []domain.User{
					{ID: "user-1", Username: "alice"},
					{ID: "user-2", Username: "bob"},
					{ID: "user-3", Username: "charlie"},
				},
			},
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Exists(ctx, "new-team").Return(false, nil).Once()
				mockTeamRepo.EXPECT().Add(ctx, domain.Team{
					Name: "new-team",
					Members: []domain.User{
						{ID: "user-1", Username: "alice"},
						{ID: "user-2", Username: "bob"},
						{ID: "user-3", Username: "charlie"},
					},
				}).Return(domain.Team{Name: "new-team"}, nil).Once()
				mockUserRepo.EXPECT().BatchExistsByID(ctx, mock.Anything).Return(map[domain.User]bool{
					{ID: "user-1", Username: "alice"}:   true,
					{ID: "user-2", Username: "bob"}:     false,
					{ID: "user-3", Username: "charlie"}: false,
				}).Once()
				mockUserRepo.EXPECT().Add(ctx, mock.MatchedBy(func(users []domain.User) bool {
					if len(users) != 2 {
						return false
					}
					ids := make(map[string]bool)
					for _, u := range users {
						ids[u.ID] = true
					}
					return ids["user-2"] && ids["user-3"]
				})).Return([]domain.User{
					{ID: "user-2", Username: "bob"},
					{ID: "user-3", Username: "charlie"},
				}, nil).Once()
			},
			wantErr: false,
			checkResult: func(result domain.Team) {
				s.Equal("new-team", result.Name)
			},
		},
		{
			name: "team already exists",
			team: domain.Team{
				Name:    "existing-team",
				Members: []domain.User{{ID: "user-1"}},
			},
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Exists(ctx, "existing-team").Return(true, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrTeamAlreadyExists,
		},
		{
			name: "exists check error",
			team: domain.Team{Name: "test-team"},
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Exists(ctx, "test-team").Return(false, errors.New("database error")).Once()
			},
			wantErr: true,
		},
		{
			name: "add team error",
			team: domain.Team{Name: "test-team"},
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Exists(ctx, "test-team").Return(false, nil).Once()
				mockTeamRepo.EXPECT().Add(ctx, domain.Team{Name: "test-team"}).Return(domain.Team{}, errors.New("insert error")).Once()
			},
			wantErr: true,
		},
		{
			name: "add users error",
			team: domain.Team{
				Name:    "test-team",
				Members: []domain.User{{ID: "user-1", Username: "alice"}},
			},
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Exists(ctx, "test-team").Return(false, nil).Once()
				mockTeamRepo.EXPECT().Add(ctx, domain.Team{
					Name:    "test-team",
					Members: []domain.User{{ID: "user-1", Username: "alice"}},
				}).Return(domain.Team{Name: "test-team"}, nil).Once()
				mockUserRepo.EXPECT().BatchExistsByID(ctx, mock.Anything).Return(map[domain.User]bool{
					{ID: "user-1", Username: "alice"}: false,
				}).Once()
				mockUserRepo.EXPECT().Add(ctx, mock.Anything).Return([]domain.User{}, errors.New("user insert error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			mockTeamRepo := newMockiTeamRepository(s.T())
			mockUserRepo := newMockiTeamUserRepository(s.T())
			service := NewTeamService(mockTeamRepo, mockUserRepo)

			tt.arrangeFunc(s.ctx, mockTeamRepo, mockUserRepo)

			// Act
			result, err := service.Add(s.ctx, tt.team)

			// Assert
			if tt.wantErr {
				s.Error(err)
				if tt.wantErrIs != nil {
					s.ErrorIs(err, tt.wantErrIs)
				}
				s.Equal(domain.Team{}, result)
			} else {
				s.NoError(err)
				if tt.checkResult != nil {
					tt.checkResult(result)
				}
			}
		})
	}
}

// TestGet проверяет метод Get
func (s *TeamServiceTestSuite) TestGet() {
	tests := []struct {
		name        string
		teamName    string
		arrangeFunc func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository)
		wantErr     bool
		wantErrIs   error
		checkResult func(result domain.Team)
	}{
		{
			name:     "success",
			teamName: "backend-team",
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Get(ctx, "backend-team").Return(domain.Team{Name: "backend-team"}, nil).Once()
				mockUserRepo.EXPECT().GetByTeamName(ctx, "backend-team").Return([]domain.User{
					{ID: "user-1", Username: "alice", TeamName: "backend-team"},
					{ID: "user-2", Username: "bob", TeamName: "backend-team"},
				}, nil).Once()
			},
			wantErr: false,
			checkResult: func(result domain.Team) {
				s.Equal("backend-team", result.Name)
				s.Len(result.Members, 2)
			},
		},
		{
			name:     "success - empty team",
			teamName: "empty-team",
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Get(ctx, "empty-team").Return(domain.Team{Name: "empty-team"}, nil).Once()
				mockUserRepo.EXPECT().GetByTeamName(ctx, "empty-team").Return([]domain.User{}, nil).Once()
			},
			wantErr: false,
			checkResult: func(result domain.Team) {
				s.Equal("empty-team", result.Name)
				s.Len(result.Members, 0)
			},
		},
		{
			name:     "team not found",
			teamName: "non-existent-team",
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Get(ctx, "non-existent-team").Return(domain.Team{}, errors.New("team not found")).Once()
			},
			wantErr: true,
		},
		{
			name:     "get members error",
			teamName: "backend-team",
			arrangeFunc: func(ctx context.Context, mockTeamRepo *mockiTeamRepository, mockUserRepo *mockiTeamUserRepository) {
				mockTeamRepo.EXPECT().Get(ctx, "backend-team").Return(domain.Team{Name: "backend-team"}, nil).Once()
				mockUserRepo.EXPECT().GetByTeamName(ctx, "backend-team").Return([]domain.User{}, errors.New("failed to get members")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			mockTeamRepo := newMockiTeamRepository(s.T())
			mockUserRepo := newMockiTeamUserRepository(s.T())
			service := NewTeamService(mockTeamRepo, mockUserRepo)

			tt.arrangeFunc(s.ctx, mockTeamRepo, mockUserRepo)

			// Act
			result, err := service.Get(s.ctx, tt.teamName)

			// Assert
			if tt.wantErr {
				s.Error(err)
				if tt.wantErrIs != nil {
					s.ErrorIs(err, tt.wantErrIs)
				}
				s.Equal(domain.Team{}, result)
			} else {
				s.NoError(err)
				if tt.checkResult != nil {
					tt.checkResult(result)
				}
			}
		})
	}
}

// TestTeamServiceSuite запускает test suite
func TestTeamServiceSuite(t *testing.T) {
	suite.Run(t, new(TeamServiceTestSuite))
}
