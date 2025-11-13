package service

type iPullRequestRepository interface {
}

type iReviewRepository interface {
}

type PullRequestService struct {
	pullRequestRepo iPullRequestRepository
	reviewRepo      iReviewRepository
}

func NewPullRequestService(pullRequestRepo iPullRequestRepository, reviewRepo iReviewRepository) *PullRequestService {
	return &PullRequestService{
		pullRequestRepo: pullRequestRepo,
		reviewRepo:      reviewRepo,
	}
}
