package queries

import "avito_test_task/internal/domain"

func (m *User) ToDomain() domain.User {
	return domain.User{
		ID:        m.ID,
		Username:  m.Username,
		TeamName:  m.TeamName,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (m *PullRequest) ToDomain() domain.PullRequest {
	var status domain.PRStatus
	// Вынес отдельным енамом на случай, если в будущем появятся другие статусы
	if m.MergedAt.IsZero() {
		status = domain.PRStatusOpen
	} else {
		status = domain.PRStatusMerged
	}
	return domain.PullRequest{
		ID:        m.ID,
		AuthorID:  m.AuthorID,
		Status:    status,
		CreatedAt: m.CreatedAt,
		MergedAt:  m.MergedAt,
	}
}

func (m *Team) ToDomain() domain.Team {
	return domain.Team{
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
