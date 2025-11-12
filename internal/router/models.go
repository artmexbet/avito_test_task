package router

import "avito_test_task/internal/domain"

type ErrorResponse struct {
	Error domain.Error `json:"error"`
}

type PullRequestShortResponse struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}
