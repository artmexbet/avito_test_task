package router

import "github.com/artmexbet/avito_test_task/internal/domain"

// ErrorResponse represents a standard error response structure.
type ErrorResponse struct {
	Error domain.Error `json:"error"`
}

// PullRequestShortResponse represents a shortened response structure for a pull request.
type PullRequestShortResponse struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}
