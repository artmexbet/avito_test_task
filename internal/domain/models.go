package domain

import "time"

// User represents a user in the system.
type User struct {
	ID        string `json:"user_id"`
	Username  string `json:"username"`
	TeamName  string `json:"team_name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PullRequest represents the status of a pull request.
type PullRequest struct {
	ID        string   `json:"pull_request_id"`
	Name      string   `json:"pull_request_name"`
	AuthorID  string   `json:"author_id"` // Появляется дублирование полей. Сделал для удобства сериализации
	Status    PRStatus `json:"status"`
	Author    *User
	Reviewers []User `json:"assigned_reviewers,omitempty"`
	CreatedAt time.Time
	MergedAt  time.Time
}

// Team represents a team in the system.
type Team struct {
	Name      string `json:"team_name"`
	Members   []User `json:"members"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Error defines the type for error codes.
type Error struct {
	Message string    `json:"message"`
	Code    ErrorCode `json:"code"`
}
