package service

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

type iPullRequestRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error)
	GetByID(ctx context.Context, prID string) (domain.PullRequest, error)
	Merge(ctx context.Context, prID string) (domain.PullRequest, error)
	Exists(ctx context.Context, prID string) (bool, error)
}

type iReviewRepository interface {
	AssignToPR(ctx context.Context, prID string, reviewerIDs []string) error
	Reassign(ctx context.Context, prID, newReviewerID, oldReviewerID string) error
	GetByPRID(ctx context.Context, prID string) ([]domain.User, error)
	GetReviewingPR(ctx context.Context, userID string) ([]domain.PullRequest, error)
}

type iPRUserRepository interface {
	GetByID(ctx context.Context, userID string) (domain.User, error)
	ExistsByID(ctx context.Context, userID string) (bool, error)
	GetActiveByTeamName(ctx context.Context, teamName string) ([]domain.User, error)
}

type PullRequestService struct {
	pullRequestRepo iPullRequestRepository
	reviewRepo      iReviewRepository
	userRepo        iPRUserRepository
}

func NewPullRequestService(
	pullRequestRepo iPullRequestRepository,
	reviewRepo iReviewRepository,
	userRepo iPRUserRepository,
) *PullRequestService {
	return &PullRequestService{
		pullRequestRepo: pullRequestRepo,
		reviewRepo:      reviewRepo,
		userRepo:        userRepo,
	}
}

func (p *PullRequestService) Create(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	// check if PR with same ID exists
	exists, err := p.pullRequestRepo.Exists(ctx, pr.ID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error checking existing pull request: %w", err)
	}
	if exists {
		return domain.PullRequest{}, fmt.Errorf("pull request with ID %s: %w", pr.ID, domain.ErrPRAlreadyExists)
	}

	// check author
	author, err := p.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error finding author: %w", err)
	}

	newPR, err := p.pullRequestRepo.Create(ctx, pr)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error creating pull request: %w", err)
	}

	// assign reviewers
	activeUsers, err := p.userRepo.GetActiveByTeamName(ctx, author.TeamName)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error getting active users by team name: %w", err)
	}
	if len(activeUsers) == 0 {
		return domain.PullRequest{},
			fmt.Errorf("no active users in team %s: %w", author.TeamName, domain.ErrNoAvailableReviewers)
	}

	// exclude author from reviewers
	var reviewerIDs []string
	for _, user := range activeUsers {
		if user.ID != author.ID {
			reviewerIDs = append(reviewerIDs, user.ID)
		}
	}
	if len(reviewerIDs) == 0 {
		return domain.PullRequest{}, fmt.Errorf("no available users to assign: %w", domain.ErrNoAvailableReviewers)
	}
	if len(reviewerIDs) > 2 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		rng.Shuffle(len(reviewerIDs), func(i, j int) {
			reviewerIDs[i], reviewerIDs[j] = reviewerIDs[j], reviewerIDs[i]
		})
		reviewerIDs = reviewerIDs[:2]
	}

	if err := p.reviewRepo.AssignToPR(ctx, newPR.ID, reviewerIDs); err != nil {
		return domain.PullRequest{}, fmt.Errorf("error assigning reviewers to pull request: %w", err)
	}

	newPR.Reviewers, err = p.reviewRepo.GetByPRID(ctx, newPR.ID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error getting author by ID: %w", err)
	}

	return newPR, nil
}

func (p *PullRequestService) Merge(ctx context.Context, prID string) (domain.PullRequest, error) {
	// check if PR exists
	pr, err := p.pullRequestRepo.GetByID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error checking existing pull request: %w", err)
	}

	// check if PR is already merged
	if pr.Status == domain.PRStatusMerged {
		return pr, fmt.Errorf("pull request with ID %s: %w", prID, domain.ErrPRAlreadyMerged)
	}

	mergedPR, err := p.pullRequestRepo.Merge(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error merging pull request: %w", err)
	}

	// retrieve reviewers
	mergedPR.Reviewers, err = p.reviewRepo.GetByPRID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("error getting reviewers for pull request: %w", err)
	}

	return mergedPR, nil
}

func (p *PullRequestService) GetReviewingPRs(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	// check user
	exists, err := p.userRepo.ExistsByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("user with ID %s: %w", userID, domain.ErrUserNotFound)
	}

	prs, err := p.reviewRepo.GetReviewingPR(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting reviewing pull requests: %w", err)
	}
	return prs, nil
}

func (p *PullRequestService) ReassignReviewer(
	ctx context.Context,
	prID, oldReviewerID string,
) (*domain.PullRequest, string, error) {
	// check if PR exists
	pr, err := p.pullRequestRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("error checking existing pull request: %w", err)
	}

	if pr.Status == domain.PRStatusMerged {
		return nil, "", fmt.Errorf("pull request with ID %s: %w", prID, domain.ErrPRAlreadyMerged)
	}

	// check old reviewer
	exists, err := p.userRepo.ExistsByID(ctx, oldReviewerID)
	if err != nil {
		return nil, "", fmt.Errorf("error checking existing user: %w", err)
	}
	if !exists {
		return nil, "", fmt.Errorf("user with ID %s: %w", oldReviewerID, domain.ErrUserNotFound)
	}

	// check if old reviewer is assigned to the PR
	assignedReviewers, err := p.reviewRepo.GetByPRID(ctx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("error getting assigned reviewers: %w", err)
	}
	mapAssignedReviewers := make(map[string]struct{}, len(assignedReviewers))
	for _, reviewer := range assignedReviewers {
		mapAssignedReviewers[reviewer.ID] = struct{}{}
	}
	if _, ok := mapAssignedReviewers[oldReviewerID]; !ok {
		return nil, "", fmt.Errorf("old reviewer with ID %s: %w", oldReviewerID, domain.ErrReviewerNotAssigned)
	}

	activeUsers, err := p.userRepo.GetActiveByTeamName(ctx, assignedReviewers[0].TeamName)
	if err != nil {
		return nil, "", fmt.Errorf("error getting active users by team name: %w", err)
	}
	r := slices.DeleteFunc(activeUsers, func(user domain.User) bool {
		_, isAssigned := mapAssignedReviewers[user.ID]
		return isAssigned || user.ID == oldReviewerID || user.ID == pr.AuthorID
	})

	if len(r) == 0 {
		return nil, "", fmt.Errorf("no available active users to reassign as reviewer: %w", domain.ErrNoAvailableReviewers)
	}

	if err := p.reviewRepo.Reassign(ctx, prID, r[0].ID, oldReviewerID); err != nil {
		return nil, "", fmt.Errorf("error reassigning reviewer: %w", err)
	}

	pr.Reviewers, err = p.reviewRepo.GetByPRID(ctx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("error getting reviewers of pull request by ID: %w", err)
	}
	return &pr, r[0].ID, nil
}
