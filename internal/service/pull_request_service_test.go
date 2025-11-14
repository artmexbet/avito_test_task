package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

// PullRequestServiceTestSuite определяет test suite для PullRequestService
type PullRequestServiceTestSuite struct {
	suite.Suite
	ctx context.Context
}

// SetupTest выполняется перед каждым тестом
func (s *PullRequestServiceTestSuite) SetupTest() {
	s.ctx = context.Background()
}

// TestCreate проверяет метод Create
func (s *PullRequestServiceTestSuite) TestCreate() {
	tests := []struct {
		name        string
		pr          domain.PullRequest
		arrangeFunc func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository)
		wantErr     bool
		wantErrIs   error
		checkResult func(result domain.PullRequest)
	}{
		{
			name: "success - multiple reviewers",
			pr: domain.PullRequest{
				ID:       "pr-1",
				Name:     "Add feature X",
				AuthorID: "author-1",
				Status:   domain.PRStatusOpen,
			},
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				author := domain.User{
					ID:       "author-1",
					Username: "alice",
					TeamName: "backend-team",
					IsActive: true,
				}
				activeUsers := []domain.User{
					{ID: "author-1", Username: "alice", TeamName: "backend-team", IsActive: true},
					{ID: "user-2", Username: "bob", TeamName: "backend-team", IsActive: true},
					{ID: "user-3", Username: "charlie", TeamName: "backend-team", IsActive: true},
				}
				createdPR := domain.PullRequest{
					ID:     "pr-1",
					Name:   "Add feature X",
					Status: domain.PRStatusOpen,
				}
				reviewers := []domain.User{
					{ID: "user-2", Username: "bob"},
					{ID: "user-3", Username: "charlie"},
				}

				mockPRRepo.EXPECT().
					Exists(ctx, "pr-1").
					Return(false, nil).Once()

				mockUserRepo.EXPECT().
					GetByID(ctx, "author-1").
					Return(author, nil).Once()

				mockPRRepo.EXPECT().
					Create(ctx, domain.PullRequest{
						ID:       "pr-1",
						Name:     "Add feature X",
						AuthorID: "author-1",
						Status:   domain.PRStatusOpen,
					}).
					Return(createdPR, nil).Once()

				mockUserRepo.EXPECT().
					GetActiveByTeamName(ctx, "backend-team").
					Return(activeUsers, nil).Once()

				mockReviewRepo.EXPECT().
					AssignToPR(ctx, "pr-1", []string{"user-2", "user-3"}).
					Return(nil).Once()

				mockReviewRepo.EXPECT().
					GetByPRID(ctx, "author-1").
					Return(reviewers, nil).Once()
			},
			wantErr: false,
			checkResult: func(result domain.PullRequest) {
				s.Equal("pr-1", result.ID)
				s.Len(result.Reviewers, 2)
			},
		},
		{
			name: "success - one reviewer",
			pr: domain.PullRequest{
				ID:       "pr-1",
				AuthorID: "author-1",
			},
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				author := domain.User{
					ID:       "author-1",
					TeamName: "small-team",
					IsActive: true,
				}
				activeUsers := []domain.User{
					{ID: "author-1", TeamName: "small-team", IsActive: true},
					{ID: "user-2", Username: "bob", TeamName: "small-team", IsActive: true},
				}
				createdPR := domain.PullRequest{ID: "pr-1"}
				reviewers := []domain.User{
					{ID: "user-2", Username: "bob"},
				}

				mockPRRepo.EXPECT().
					Exists(ctx, "pr-1").
					Return(false, nil).Once()

				mockUserRepo.EXPECT().
					GetByID(ctx, "author-1").
					Return(author, nil).Once()

				mockPRRepo.EXPECT().
					Create(ctx, domain.PullRequest{
						ID:       "pr-1",
						AuthorID: "author-1",
					}).
					Return(createdPR, nil).Once()

				mockUserRepo.EXPECT().
					GetActiveByTeamName(ctx, "small-team").
					Return(activeUsers, nil).Once()

				mockReviewRepo.EXPECT().
					AssignToPR(ctx, "pr-1", []string{"user-2"}).
					Return(nil).Once()

				mockReviewRepo.EXPECT().
					GetByPRID(ctx, "author-1").
					Return(reviewers, nil).Once()
			},
			wantErr: false,
			checkResult: func(result domain.PullRequest) {
				s.Equal("pr-1", result.ID)
				s.Len(result.Reviewers, 1)
			},
		},
		{
			name: "PR already exists",
			pr: domain.PullRequest{
				ID:       "existing-pr",
				AuthorID: "author-1",
			},
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				mockPRRepo.EXPECT().
					Exists(ctx, "existing-pr").
					Return(true, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrPRAlreadyExists,
		},
		{
			name: "exists check error",
			pr:   domain.PullRequest{ID: "pr-1"},
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				mockPRRepo.EXPECT().
					Exists(ctx, "pr-1").
					Return(false, errors.New("database error")).Once()
			},
			wantErr: true,
		},
		{
			name: "author not found",
			pr: domain.PullRequest{
				ID:       "pr-1",
				AuthorID: "non-existent",
			},
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				mockPRRepo.EXPECT().
					Exists(ctx, "pr-1").
					Return(false, nil).Once()

				mockUserRepo.EXPECT().
					GetByID(ctx, "non-existent").
					Return(domain.User{}, errors.New("author not found")).Once()
			},
			wantErr: true,
		},
		{
			name: "create error",
			pr: domain.PullRequest{
				ID:       "pr-1",
				AuthorID: "author-1",
			},
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				author := domain.User{ID: "author-1", TeamName: "team-1"}
				mockPRRepo.EXPECT().
					Exists(ctx, "pr-1").
					Return(false, nil).Once()

				mockUserRepo.EXPECT().
					GetByID(ctx, "author-1").
					Return(author, nil).Once()

				mockPRRepo.EXPECT().
					Create(ctx, domain.PullRequest{
						ID:       "pr-1",
						AuthorID: "author-1",
					}).
					Return(domain.PullRequest{}, errors.New("insert error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			mockPRRepo := newMockiPullRequestRepository(s.T())
			mockReviewRepo := newMockiReviewRepository(s.T())
			mockUserRepo := newMockiPRUserRepository(s.T())
			service := NewPullRequestService(mockPRRepo, mockReviewRepo, mockUserRepo)

			tt.arrangeFunc(s.ctx, mockPRRepo, mockReviewRepo, mockUserRepo)

			// Act
			result, err := service.Create(s.ctx, tt.pr)

			// Assert
			if tt.wantErr {
				s.Error(err)
				if tt.wantErrIs != nil {
					s.ErrorIs(err, tt.wantErrIs)
				}
				s.Equal(domain.PullRequest{}, result)
			} else {
				s.NoError(err)
				if tt.checkResult != nil {
					tt.checkResult(result)
				}
			}
		})
	}
}

// TestMerge проверяет метод Merge
func (s *PullRequestServiceTestSuite) TestMerge() {
	tests := []struct {
		name        string
		prID        string
		arrangeFunc func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository)
		wantErr     bool
		wantErrIs   error
		checkResult func(result domain.PullRequest)
	}{
		{
			name: "success",
			prID: "pr-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository) {
				mergedPR := domain.PullRequest{
					ID:     "pr-1",
					Status: domain.PRStatusMerged,
				}
				reviewers := []domain.User{
					{ID: "user-1", Username: "alice"},
					{ID: "user-2", Username: "bob"},
				}
				mockPRRepo.EXPECT().GetByID(ctx, "pr-1").Return(domain.PullRequest{}, nil).Once()
				mockPRRepo.EXPECT().Merge(ctx, "pr-1").Return(mergedPR, nil).Once()
				mockReviewRepo.EXPECT().GetByPRID(ctx, "pr-1").Return(reviewers, nil).Once()
			},
			wantErr: false,
			checkResult: func(result domain.PullRequest) {
				s.Equal("pr-1", result.ID)
				s.Equal(domain.PRStatusMerged, result.Status)
			},
		},
		{
			name: "PR not found",
			prID: "non-existent-pr",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository) {
				mockPRRepo.EXPECT().GetByID(ctx, "non-existent-pr").Return(domain.PullRequest{}, domain.ErrPRNotFound).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrPRNotFound,
		},
		{
			name: "exists check error",
			prID: "pr-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository) {
				mockPRRepo.EXPECT().GetByID(ctx, "pr-1").Return(domain.PullRequest{}, domain.ErrPRNotFound).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrPRNotFound,
		},
		{
			name: "merge error",
			prID: "pr-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository) {
				mockPRRepo.EXPECT().GetByID(ctx, "pr-1").Return(domain.PullRequest{}, nil).Once()
				mockPRRepo.EXPECT().Merge(ctx, "pr-1").Return(domain.PullRequest{}, errors.New("merge failed")).Once()
			},
			wantErr: true,
		},
		{
			name: "get reviewers error",
			prID: "pr-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository) {
				mergedPR := domain.PullRequest{ID: "pr-1", Status: domain.PRStatusMerged}
				mockPRRepo.EXPECT().GetByID(ctx, "pr-1").Return(domain.PullRequest{}, nil).Once()
				mockPRRepo.EXPECT().Merge(ctx, "pr-1").Return(mergedPR, nil).Once()
				mockReviewRepo.EXPECT().GetByPRID(ctx, "pr-1").Return([]domain.User{}, errors.New("failed to get reviewers")).Once()
			},
			wantErr: true,
		},
		{
			name: "pr already merged",
			prID: "pr-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository) {
				mockPRRepo.EXPECT().GetByID(ctx, "pr-1").Return(domain.PullRequest{
					Status: domain.PRStatusMerged,
				}, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrPRAlreadyMerged,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			mockPRRepo := newMockiPullRequestRepository(s.T())
			mockReviewRepo := newMockiReviewRepository(s.T())
			mockUserRepo := newMockiPRUserRepository(s.T())
			service := NewPullRequestService(mockPRRepo, mockReviewRepo, mockUserRepo)

			tt.arrangeFunc(s.ctx, mockPRRepo, mockReviewRepo)

			// Act
			result, err := service.Merge(s.ctx, tt.prID)

			// Assert
			if tt.wantErr {
				s.Error(err)
				if tt.wantErrIs != nil {
					s.ErrorIs(err, tt.wantErrIs)
				}
				s.Equal(domain.PullRequest{}, result)
			} else {
				s.NoError(err)
				if tt.checkResult != nil {
					tt.checkResult(result)
				}
			}
		})
	}
}

// TestGetReviewingPRs проверяет метод GetReviewingPRs
func (s *PullRequestServiceTestSuite) TestGetReviewingPRs() {
	tests := []struct {
		name        string
		userID      string
		arrangeFunc func(ctx context.Context, mockUserRepo *mockiPRUserRepository, mockReviewRepo *mockiReviewRepository)
		wantErr     bool
		wantErrIs   error
		checkResult func(result []domain.PullRequest)
	}{
		{
			name:   "success - multiple PRs",
			userID: "user-1",
			arrangeFunc: func(ctx context.Context, mockUserRepo *mockiPRUserRepository, mockReviewRepo *mockiReviewRepository) {
				prs := []domain.PullRequest{
					{ID: "pr-1", Name: "Feature A", Status: domain.PRStatusOpen},
					{ID: "pr-2", Name: "Feature B", Status: domain.PRStatusOpen},
				}
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-1").Return(true, nil).Once()
				mockReviewRepo.EXPECT().GetReviewingPR(ctx, "user-1").Return(prs, nil).Once()
			},
			wantErr: false,
			checkResult: func(result []domain.PullRequest) {
				s.Len(result, 2)
			},
		},
		{
			name:   "success - empty list",
			userID: "user-1",
			arrangeFunc: func(ctx context.Context, mockUserRepo *mockiPRUserRepository, mockReviewRepo *mockiReviewRepository) {
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-1").Return(true, nil).Once()
				mockReviewRepo.EXPECT().GetReviewingPR(ctx, "user-1").Return([]domain.PullRequest{}, nil).Once()
			},
			wantErr: false,
			checkResult: func(result []domain.PullRequest) {
				s.Len(result, 0)
			},
		},
		{
			name:   "user not found",
			userID: "non-existent-user",
			arrangeFunc: func(ctx context.Context, mockUserRepo *mockiPRUserRepository, mockReviewRepo *mockiReviewRepository) {
				mockUserRepo.EXPECT().ExistsByID(ctx, "non-existent-user").Return(false, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrUserNotFound,
		},
		{
			name:   "exists check error",
			userID: "user-1",
			arrangeFunc: func(ctx context.Context, mockUserRepo *mockiPRUserRepository, mockReviewRepo *mockiReviewRepository) {
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-1").Return(false, errors.New("database error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			mockPRRepo := newMockiPullRequestRepository(s.T())
			mockReviewRepo := newMockiReviewRepository(s.T())
			mockUserRepo := newMockiPRUserRepository(s.T())
			service := NewPullRequestService(mockPRRepo, mockReviewRepo, mockUserRepo)

			tt.arrangeFunc(s.ctx, mockUserRepo, mockReviewRepo)

			// Act
			result, err := service.GetReviewingPRs(s.ctx, tt.userID)

			// Assert
			if tt.wantErr {
				s.Error(err)
				if tt.wantErrIs != nil {
					s.ErrorIs(err, tt.wantErrIs)
				}
				s.Nil(result)
			} else {
				s.NoError(err)
				if tt.checkResult != nil {
					tt.checkResult(result)
				}
			}
		})
	}
}

// TestReassignReviewer проверяет метод ReassignReviewer
func (s *PullRequestServiceTestSuite) TestReassignReviewer() {
	tests := []struct {
		name          string
		prID          string
		oldReviewerID string
		arrangeFunc   func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository)
		wantErr       bool
		wantErrIs     error
		checkResult   func(result *domain.PullRequest, newID string)
	}{
		{
			name:          "success",
			prID:          "pr-1",
			oldReviewerID: "user-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				assignedReviewers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "backend-team"},
					{ID: "user-2", Username: "bob", TeamName: "backend-team"},
				}
				activeUsers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "backend-team", IsActive: true},
					{ID: "user-2", Username: "bob", TeamName: "backend-team", IsActive: true},
					{ID: "user-3", Username: "charlie", TeamName: "backend-team", IsActive: true},
				}
				updatedPR := domain.PullRequest{
					ID:     "pr-1",
					Status: domain.PRStatusOpen,
				}

				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(true, nil).Once()
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-1").Return(true, nil).Once()
				mockReviewRepo.EXPECT().GetByPRID(ctx, "pr-1").Return(assignedReviewers, nil).Once()
				mockUserRepo.EXPECT().GetActiveByTeamName(ctx, "backend-team").Return(activeUsers, nil).Once()
				mockReviewRepo.EXPECT().Reassign(ctx, "pr-1", "user-3", "user-1").Return(nil).Once()
				mockPRRepo.EXPECT().GetByID(ctx, "pr-1").Return(updatedPR, nil).Once()
			},
			wantErr: false,
			checkResult: func(result *domain.PullRequest, newID string) {
				s.NotNil(result)
				s.Equal("pr-1", result.ID)
				s.Equal("user-3", newID)
			},
		},
		{
			name:          "PR not found",
			prID:          "non-existent-pr",
			oldReviewerID: "user-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				mockPRRepo.EXPECT().Exists(ctx, "non-existent-pr").Return(false, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrPRNotFound,
		},
		{
			name:          "old reviewer not found",
			prID:          "pr-1",
			oldReviewerID: "non-existent-user",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(true, nil).Once()
				mockUserRepo.EXPECT().ExistsByID(ctx, "non-existent-user").Return(false, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrUserNotFound,
		},
		{
			name:          "reviewer not assigned",
			prID:          "pr-1",
			oldReviewerID: "user-3",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				assignedReviewers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "backend-team"},
					{ID: "user-2", Username: "bob", TeamName: "backend-team"},
				}
				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(true, nil).Once()
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-3").Return(true, nil).Once()
				mockReviewRepo.EXPECT().GetByPRID(ctx, "pr-1").Return(assignedReviewers, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrReviewerNotAssigned,
		},
		{
			name:          "no available reviewers",
			prID:          "pr-1",
			oldReviewerID: "user-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				assignedReviewers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "small-team"},
					{ID: "user-2", Username: "bob", TeamName: "small-team"},
				}
				activeUsers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "small-team", IsActive: true},
					{ID: "user-2", Username: "bob", TeamName: "small-team", IsActive: true},
				}
				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(true, nil).Once()
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-1").Return(true, nil).Once()
				mockReviewRepo.EXPECT().GetByPRID(ctx, "pr-1").Return(assignedReviewers, nil).Once()
				mockUserRepo.EXPECT().GetActiveByTeamName(ctx, "small-team").Return(activeUsers, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrNoAvailableReviewers,
		},
		{
			name:          "reassign error",
			prID:          "pr-1",
			oldReviewerID: "user-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				assignedReviewers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "backend-team"},
					{ID: "user-2", Username: "bob", TeamName: "backend-team"},
				}
				activeUsers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "backend-team", IsActive: true},
					{ID: "user-2", Username: "bob", TeamName: "backend-team", IsActive: true},
					{ID: "user-3", Username: "charlie", TeamName: "backend-team", IsActive: true},
				}
				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(true, nil).Once()
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-1").Return(true, nil).Once()
				mockReviewRepo.EXPECT().GetByPRID(ctx, "pr-1").Return(assignedReviewers, nil).Once()
				mockUserRepo.EXPECT().GetActiveByTeamName(ctx, "backend-team").Return(activeUsers, nil).Once()
				mockReviewRepo.EXPECT().Reassign(ctx, "pr-1", "user-3", "user-1").Return(errors.New("reassign failed")).Once()
			},
			wantErr: true,
		},
		{
			name:          "get PR by ID error",
			prID:          "pr-1",
			oldReviewerID: "user-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				assignedReviewers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "backend-team"},
					{ID: "user-2", Username: "bob", TeamName: "backend-team"},
				}
				activeUsers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "backend-team", IsActive: true},
					{ID: "user-2", Username: "bob", TeamName: "backend-team", IsActive: true},
					{ID: "user-3", Username: "charlie", TeamName: "backend-team", IsActive: true},
				}
				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(true, nil).Once()
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-1").Return(true, nil).Once()
				mockReviewRepo.EXPECT().GetByPRID(ctx, "pr-1").Return(assignedReviewers, nil).Once()
				mockUserRepo.EXPECT().GetActiveByTeamName(ctx, "backend-team").Return(activeUsers, nil).Once()
				mockReviewRepo.EXPECT().Reassign(ctx, "pr-1", "user-3", "user-1").Return(nil).Once()
				mockPRRepo.EXPECT().GetByID(ctx, "pr-1").Return(domain.PullRequest{}, errors.New("failed to get PR")).Once()
			},
			wantErr: true,
		},
		{
			name:          "cannot reassign, not enough active users",
			prID:          "pr-1",
			oldReviewerID: "user-1",
			arrangeFunc: func(ctx context.Context, mockPRRepo *mockiPullRequestRepository, mockReviewRepo *mockiReviewRepository, mockUserRepo *mockiPRUserRepository) {
				assignedReviewers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "solo-team"},
				}
				activeUsers := []domain.User{
					{ID: "user-1", Username: "alice", TeamName: "solo-team", IsActive: true},
				}
				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(true, nil).Once()
				mockUserRepo.EXPECT().ExistsByID(ctx, "user-1").Return(true, nil).Once()
				mockReviewRepo.EXPECT().GetByPRID(ctx, "pr-1").Return(assignedReviewers, nil).Once()
				mockUserRepo.EXPECT().GetActiveByTeamName(ctx, "solo-team").Return(activeUsers, nil).Once()
			},
			wantErr:   true,
			wantErrIs: domain.ErrNoAvailableReviewers,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			mockPRRepo := newMockiPullRequestRepository(s.T())
			mockReviewRepo := newMockiReviewRepository(s.T())
			mockUserRepo := newMockiPRUserRepository(s.T())
			service := NewPullRequestService(mockPRRepo, mockReviewRepo, mockUserRepo)

			tt.arrangeFunc(s.ctx, mockPRRepo, mockReviewRepo, mockUserRepo)

			// Act
			result, newID, err := service.ReassignReviewer(s.ctx, tt.prID, tt.oldReviewerID)

			// Assert
			if tt.wantErr {
				s.Error(err)
				if tt.wantErrIs != nil {
					s.ErrorIs(err, tt.wantErrIs)
				}
				s.Nil(result)
				s.Empty(newID)
			} else {
				s.NoError(err)
				if tt.checkResult != nil {
					tt.checkResult(result, newID)
				}
			}
		})
	}
}

// TestPullRequestServiceSuite запускает test suite
func TestPullRequestServiceSuite(t *testing.T) {
	suite.Run(t, new(PullRequestServiceTestSuite))
}
