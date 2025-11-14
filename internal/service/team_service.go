package service

import (
	"context"
	"fmt"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

type iTeamRepository interface {
	Get(ctx context.Context, teamName string) (domain.Team, error)
	Add(ctx context.Context, team domain.Team) (domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

type iTeamUserRepository interface {
	Add(ctx context.Context, users []domain.User) ([]domain.User, error)
	BatchExistsByID(ctx context.Context, users []domain.User) map[domain.User]bool
	GetByTeamName(ctx context.Context, teamName string) ([]domain.User, error)
}

type TeamService struct {
	repository     iTeamRepository
	userRepository iTeamUserRepository
}

func NewTeamService(repository iTeamRepository, userRepository iTeamUserRepository) *TeamService {
	return &TeamService{
		repository:     repository,
		userRepository: userRepository,
	}
}

func (s *TeamService) Add(ctx context.Context, team domain.Team) (domain.Team, error) {
	exists, err := s.repository.Exists(ctx, team.Name)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to check if team exists by name %s: %w", team.Name, err)
	}
	if exists {
		return domain.Team{}, fmt.Errorf("team with name %s: %w", team.Name, domain.ErrTeamAlreadyExists)
	}

	teamFromDB, err := s.repository.Add(ctx, team)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to add team to repository: %w", err)
	}
	teamFromDB.Members = team.Members

	// Check if any user is not exist
	userExistMap := s.userRepository.BatchExistsByID(ctx, team.Members)
	var usersToAdd []domain.User
	for _, user := range team.Members {
		if !userExistMap[user] {
			usersToAdd = append(usersToAdd, user)
		}
	}

	if len(usersToAdd) == 0 {
		return teamFromDB, nil
	}

	// Add non-existing users
	_, err = s.userRepository.Add(ctx, usersToAdd)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to add team members to user repository: %w", err)
	}

	return teamFromDB, nil
}

func (s *TeamService) Get(ctx context.Context, teamName string) (domain.Team, error) {
	team, err := s.repository.Get(ctx, teamName)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to get team by name %s: %w", teamName, err)
	}

	team.Members, err = s.userRepository.GetByTeamName(ctx, teamName)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to get team members by team name %s: %w", teamName, err)
	}

	return team, nil
}
