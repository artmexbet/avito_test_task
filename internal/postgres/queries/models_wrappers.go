package queries

import "github.com/artmexbet/avito_test_task/internal/domain"

// ToDomain converts the User model to the domain User model.
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

// ToDomain converts the PullRequest model to the domain PullRequest model.
func (m *PullRequest) ToDomain() domain.PullRequest {
	var status domain.PRStatus
	// Вынес отдельным енамом на случай, если в будущем появятся другие статусы
	if m.MergedAt.IsZero() {
		status = domain.PRStatusOpen
	} else {
		status = domain.PRStatusMerged
	}
	return domain.PullRequest{ //nolint:exhaustruct // Не все доменные поля можно заполнить отсюда
		ID:        m.ID,
		AuthorID:  m.AuthorID,
		Status:    status,
		CreatedAt: m.CreatedAt,
		MergedAt:  m.MergedAt,
	}
}

// ToDomain converts the Team model to the domain Team model.
func (m *Team) ToDomain() domain.Team {
	return domain.Team{ //nolint:exhaustruct // Не все доменные поля можно заполнить отсюда
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
