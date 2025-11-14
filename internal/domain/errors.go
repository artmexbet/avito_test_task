package domain

import "errors"

var (
	ErrTeamAlreadyExists    = errors.New("team already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrPRAlreadyExists      = errors.New("pull request already exists")
	ErrPRNotFound           = errors.New("pull request not found")
	ErrReviewerNotAssigned  = errors.New("reviewer not assigned to the pull request")
	ErrNoAvailableReviewers = errors.New("no available reviewers to assign")
)
