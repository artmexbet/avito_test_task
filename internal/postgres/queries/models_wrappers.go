package queries

import (
	"time"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

// ToDomain converts the User model to the domain User model.
func (m *User) ToDomain() domain.User {
	var updatedAt time.Time
	if m.UpdatedAt != nil {
		updatedAt = *m.UpdatedAt
	}
	return domain.User{
		ID:        m.ID,
		Username:  m.Username,
		TeamName:  m.TeamName,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: updatedAt,
	}
}

// ToDomain converts the PullRequest model to the domain PullRequest model.
func (m *PullRequest) ToDomain() domain.PullRequest {
	var status domain.PRStatus
	var mergedAt time.Time
	// Вынес отдельным енамом на случай, если в будущем появятся другие статусы
	if m.MergedAt == nil {
		status = domain.PRStatusOpen
	} else {
		status = domain.PRStatusMerged
		mergedAt = *m.MergedAt
	}
	return domain.PullRequest{ //nolint:exhaustruct // Не все доменные поля можно заполнить отсюда
		ID:        m.ID,
		Name:      m.Name,
		AuthorID:  m.AuthorID,
		Status:    status,
		CreatedAt: m.CreatedAt,
		MergedAt:  mergedAt,
	}
}

// ToDomain converts the Team model to the domain Team model.
func (m *Team) ToDomain() domain.Team {
	var updatedAt time.Time
	if m.UpdatedAt != nil {
		updatedAt = *m.UpdatedAt
	}
	return domain.Team{ //nolint:exhaustruct // Не все доменные поля можно заполнить отсюда
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: updatedAt,
	}
}
