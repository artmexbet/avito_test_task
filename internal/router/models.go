package router

import (
	"time"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

// При использовании не с Fiber можно генерировать анмаршаллер //go:generate easyjson -all models.go

// ErrorCode defines the type for error codes.
type ErrorCode string

// Possible values for ErrorCode
const (
	errorCodeTeamExist ErrorCode = "TEAM_EXISTS"
	errorCodeNotFound  ErrorCode = "NOT_FOUND"
	// PullRequest specific error codes
	errorCodePRExists    ErrorCode = "PR_EXISTS"
	errorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	errorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
)

// Error defines the type for error codes.
type Error struct {
	Message string    `json:"message"`
	Code    ErrorCode `json:"code"`
}

// errorResponse represents a standard error response structure.
type errorResponse struct {
	Error Error `json:"error"`
}

var (
	errorResponseNotFound = errorResponse{
		Error: Error{
			Message: "resource not found",
			Code:    errorCodeNotFound,
		},
	}

	errorBadRequest = errorResponse{
		Error: Error{
			Message: "bad request",
			Code:    "BAD_REQUEST",
		},
	}
)

func newErrorResponse(message string, code ErrorCode) errorResponse {
	return errorResponse{
		Error: Error{
			Message: message,
			Code:    code,
		},
	}
}

type pullRequestResponse struct {
	ID        string          `json:"pull_request_id"`
	Name      string          `json:"pull_request_name"`
	AuthorID  string          `json:"author_id"`
	Reviewers []string        `json:"assigned_reviewers,omitempty"`
	Status    domain.PRStatus `json:"status"`
	MergedAt  time.Time       `json:"merged_at"`
}

// pullRequestShortResponse represents a shortened response structure for a pull request.
type pullRequestShortResponse struct {
	ID       string          `json:"pull_request_id"`
	Name     string          `json:"pull_request_name"`
	AuthorID string          `json:"author_id"`
	Status   domain.PRStatus `json:"status"`
}

// fromDomainPR converts domain.PullRequest to pullRequestResponse
func fromDomainPR(pr domain.PullRequest) pullRequestResponse {
	resp := pullRequestResponse{
		ID:        pr.ID,
		Name:      pr.Name,
		AuthorID:  pr.AuthorID,
		Reviewers: make([]string, 0, len(pr.Reviewers)),
		Status:    pr.Status,
		MergedAt:  pr.MergedAt,
	}
	if len(pr.Reviewers) > 0 {
		resp.Reviewers = make([]string, 0, len(pr.Reviewers))
		for _, r := range pr.Reviewers {
			resp.Reviewers = append(resp.Reviewers, r.ID)
		}
	}
	return resp
}

type member struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active" validate:"-"`
}

type addTeamRequest struct {
	TeamName string   `json:"team_name" validate:"required"`
	Members  []member `json:"members" validate:"required,dive"`
}

func (r *addTeamRequest) ToDomain() domain.Team {
	var members []domain.User
	for _, m := range r.Members {
		members = append(members, domain.User{ //nolint:exhaustruct
			ID:       m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
			TeamName: r.TeamName,
		})
	}
	return domain.Team{ //nolint:exhaustruct
		Name:    r.TeamName,
		Members: members,
	}
}

type getTeamResponse struct {
	TeamName string   `json:"team_name"`
	Members  []member `json:"members"`
}

// fromDomainTeam converts domain.Team to getTeamResponse
func fromDomainTeam(team domain.Team) getTeamResponse {
	var members []member
	for _, m := range team.Members {
		members = append(members, member{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	return getTeamResponse{
		TeamName: team.Name,
		Members:  members,
	}
}

type setUserIsActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type reviewPRsResponse struct {
	UserID       string                     `json:"user_id"`
	PullRequests []pullRequestShortResponse `json:"pull_requests"`
}

// PullRequests requests/responses

type createPRRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

func (r createPRRequest) ToDomain() domain.PullRequest {
	return domain.PullRequest{ //nolint:exhaustruct
		ID:       r.PullRequestID,
		Name:     r.PullRequestName,
		AuthorID: r.AuthorID,
		Status:   domain.PRStatusOpen,
	}
}

type mergePRRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

type reassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldUserID     string `json:"old_user_id" validate:"required"`
}

type reassignReviewerResponse struct {
	PR         pullRequestResponse `json:"pr"`
	ReplacedBy string              `json:"replaced_by"`
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

func fromDomainUser(user domain.User) UserResponse {
	return UserResponse{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}
