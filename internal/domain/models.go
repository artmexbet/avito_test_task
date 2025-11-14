package domain

import "time"

// User represents a user in the system.
type User struct {
	ID        string
	Username  string
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PullRequest represents the status of a pull request.
type PullRequest struct {
	ID        string
	Name      string
	AuthorID  string
	Status    PRStatus
	Author    *User // Not mapped to DB
	Reviewers []User
	CreatedAt time.Time
	MergedAt  time.Time
}

// Team represents a team in the system.
type Team struct {
	Name      string
	Members   []User
	CreatedAt time.Time
	UpdatedAt time.Time
}
