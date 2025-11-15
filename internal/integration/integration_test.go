package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/artmexbet/avito_test_task/internal/domain"
	postgresRepo "github.com/artmexbet/avito_test_task/internal/postgres"
	"github.com/artmexbet/avito_test_task/internal/repository"
	"github.com/artmexbet/avito_test_task/internal/service"
)

// IntegrationTestSuite определяет test suite для интеграционных тестов
type IntegrationTestSuite struct {
	suite.Suite
	ctx           context.Context
	pgContainer   *postgres.PostgresContainer
	pool          *pgxpool.Pool
	prService     *service.PullRequestService
	userService   *service.UserService
	teamService   *service.TeamService
	userRepo      *repository.UserRepository
	prRepo        *repository.PRRepository
	reviewersRepo *repository.ReviewersRepository
	teamRepo      *repository.TeamRepository
}

// SetupSuite выполняется один раз перед всеми тестами
func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Получаем путь к директории с миграциями
	migrationsPath, err := filepath.Abs("../../migrations")
	s.Require().NoError(err)

	// Создаем PostgreSQL контейнер
	pgContainer, err := postgres.Run(s.ctx,
		"postgres:latest",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts(
			filepath.Join(migrationsPath, "01_teams.up.sql"),
			filepath.Join(migrationsPath, "02_users.up.sql"),
			filepath.Join(migrationsPath, "03_pull_requests.up.sql"),
			filepath.Join(migrationsPath, "04_pull_requests_reviewers.up.sql"),
		),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(120*time.Second), // Увеличиваем timeout для Windows
		),
	)
	s.Require().NoError(err)
	s.pgContainer = pgContainer

	// Получаем connection string
	connStr, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	// Создаем connection pool
	pool, err := pgxpool.New(s.ctx, connStr)
	s.Require().NoError(err)
	s.pool = pool

	// Инициализируем репозитории
	pg := postgresRepo.New(pool)
	s.userRepo = repository.NewUserRepository(pg)
	s.reviewersRepo = repository.NewReviewersRepository(pg)
	s.prRepo = repository.NewPRRepository(pg)
	s.teamRepo = repository.NewTeamRepository(pg)

	// Инициализируем сервисы
	s.prService = service.NewPullRequestService(s.prRepo, s.reviewersRepo, s.userRepo)
	s.userService = service.NewUserService(s.userRepo)
	s.teamService = service.NewTeamService(s.teamRepo, s.userRepo)
}

// TearDownSuite выполняется один раз после всех тестов
func (s *IntegrationTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.pgContainer != nil {
		err := s.pgContainer.Terminate(s.ctx)
		s.Require().NoError(err)
	}
}

// SetupTest выполняется перед каждым тестом
func (s *IntegrationTestSuite) SetupTest() {
	// Очищаем все таблицы перед каждым тестом
	_, err := s.pool.Exec(s.ctx, "TRUNCATE TABLE pull_requests_reviewers, pull_requests, users, teams CASCADE")
	s.Require().NoError(err)
}

// TestFullPRWorkflow тестирует полный workflow создания PR, назначения ревьюверов и мерджа
func (s *IntegrationTestSuite) TestFullPRWorkflow() {
	// Создаем команду с пользователями
	team := domain.Team{
		Name: "backend-team",
		Members: []domain.User{
			{ID: "user-1", Username: "alice", TeamName: "backend-team", IsActive: true},
			{ID: "user-2", Username: "bob", TeamName: "backend-team", IsActive: true},
			{ID: "user-3", Username: "charlie", TeamName: "backend-team", IsActive: true},
		},
	}

	createdTeam, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)
	s.Equal("backend-team", createdTeam.Name)
	s.Len(createdTeam.Members, 3)

	// Создаем PR от имени user-1
	pr := domain.PullRequest{
		ID:       "pr-1",
		Name:     "Add feature X",
		AuthorID: "user-1",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)

	s.Equal("pr-1", createdPR.ID)
	s.Equal("Add feature X", createdPR.Name)
	s.Equal("user-1", createdPR.AuthorID)
	s.Equal(domain.PRStatusOpen, createdPR.Status)
	s.Len(createdPR.Reviewers, 2) // Должно быть 2 ревьювера (не включая автора)

	// Проверяем, что автор не назначен ревьювером
	for _, reviewer := range createdPR.Reviewers {
		s.NotEqual("user-1", reviewer.ID)
	}

	// Проверяем, что у ревьювера есть этот PR в списке на ревью
	reviewingPRs, err := s.prService.GetReviewingPRs(s.ctx, createdPR.Reviewers[0].ID)
	s.Require().NoError(err)
	s.Len(reviewingPRs, 1)
	s.Equal("pr-1", reviewingPRs[0].ID)

	// Мерджим PR
	mergedPR, err := s.prService.Merge(s.ctx, "pr-1")
	s.Require().NoError(err)
	s.Equal(domain.PRStatusMerged, mergedPR.Status)
	s.NotNil(mergedPR.MergedAt)

	// Попытка мерджа уже смердженного PR должна вернуть ошибку
	_, err = s.prService.Merge(s.ctx, "pr-1")
	s.Error(err)
	s.ErrorIs(err, domain.ErrPRAlreadyMerged)
}

// TestReassignReviewer тестирует переназначение ревьювера
func (s *IntegrationTestSuite) TestReassignReviewer() {
	// Создаем команду с пользователями
	team := domain.Team{
		Name: "frontend-team",
		Members: []domain.User{
			{ID: "user-10", Username: "alice", TeamName: "frontend-team", IsActive: true},
			{ID: "user-11", Username: "bob", TeamName: "frontend-team", IsActive: true},
			{ID: "user-12", Username: "charlie", TeamName: "frontend-team", IsActive: true},
			{ID: "user-13", Username: "diana", TeamName: "frontend-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR
	pr := domain.PullRequest{
		ID:       "pr-10",
		Name:     "Fix bug Y",
		AuthorID: "user-10",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)
	s.Len(createdPR.Reviewers, 2)

	// Сохраняем ID одного из ревьюверов для переназначения
	oldReviewerID := createdPR.Reviewers[0].ID

	// Переназначаем ревьювера
	reassignedPR, newReviewerID, err := s.prService.ReassignReviewer(s.ctx, "pr-10", oldReviewerID)
	s.Require().NoError(err)
	s.NotNil(reassignedPR)
	s.NotEmpty(newReviewerID)

	// Проверяем, что новый ревьювер не совпадает со старым
	s.NotEqual(oldReviewerID, newReviewerID)

	// Проверяем, что новый ревьювер не является автором
	s.NotEqual("user-10", newReviewerID)

	// Проверяем, что новый ревьювер не совпадает с другим ревьювером
	reviewers, err := s.reviewersRepo.GetByPRID(s.ctx, reassignedPR.ID)
	s.Require().NoError(err)
	s.Require().Len(reviewers, 2)
	s.Require().NotEqual(reviewers[0], reviewers[1])

	// Проверяем, что у старого ревьювера PR больше нет в списке на ревью
	oldReviewerPRs, err := s.prService.GetReviewingPRs(s.ctx, oldReviewerID)
	s.Require().NoError(err)
	for _, pr := range oldReviewerPRs {
		s.NotEqual("pr-10", pr.ID)
	}

	// Проверяем, что у нового ревьювера PR есть в списке на ревью
	newReviewerPRs, err := s.prService.GetReviewingPRs(s.ctx, newReviewerID)
	s.Require().NoError(err)
	found := false
	for _, pr := range newReviewerPRs {
		if pr.ID == "pr-10" {
			found = true
			break
		}
	}
	s.True(found)
}

// TestUserActivationDeactivation тестирует активацию и деактивацию пользователя
func (s *IntegrationTestSuite) TestUserActivationDeactivation() {
	// Создаем команду с пользователями
	team := domain.Team{
		Name: "devops-team",
		Members: []domain.User{
			{ID: "user-20", Username: "alice", TeamName: "devops-team", IsActive: true},
			{ID: "user-21", Username: "bob", TeamName: "devops-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Деактивируем пользователя
	updatedUser, err := s.userService.SetIsActive(s.ctx, "user-21", false)
	s.Require().NoError(err)
	s.False(updatedUser.IsActive)

	// Активируем пользователя обратно
	updatedUser, err = s.userService.SetIsActive(s.ctx, "user-21", true)
	s.Require().NoError(err)
	s.True(updatedUser.IsActive)
}

// TestSmallTeamPRCreation тестирует создание PR в малой команде (2 человека)
func (s *IntegrationTestSuite) TestSmallTeamPRCreation() {
	// Создаем команду с 2 пользователями
	team := domain.Team{
		Name: "small-team",
		Members: []domain.User{
			{ID: "user-30", Username: "alice", TeamName: "small-team", IsActive: true},
			{ID: "user-31", Username: "bob", TeamName: "small-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR от имени user-30
	pr := domain.PullRequest{
		ID:       "pr-30",
		Name:     "Small team PR",
		AuthorID: "user-30",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)
	s.Len(createdPR.Reviewers, 1) // Должен быть только 1 ревьювер
	s.Equal("user-31", createdPR.Reviewers[0].ID)
}

// TestNoAvailableReviewers тестирует сценарий, когда нет доступных ревьюверов для переназначения
func (s *IntegrationTestSuite) TestNoAvailableReviewers() {
	// Создаем команду с 2 пользователями
	team := domain.Team{
		Name: "tiny-team",
		Members: []domain.User{
			{ID: "user-40", Username: "alice", TeamName: "tiny-team", IsActive: true},
			{ID: "user-41", Username: "bob", TeamName: "tiny-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR от имени user-40
	pr := domain.PullRequest{
		ID:       "pr-40",
		Name:     "Tiny team PR",
		AuthorID: "user-40",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)
	s.Len(createdPR.Reviewers, 1)
	s.Equal("user-41", createdPR.Reviewers[0].ID)

	// Попытка переназначения должна провалиться, так как нет других кандидатов
	_, _, err = s.prService.ReassignReviewer(s.ctx, "pr-40", "user-41")
	s.Error(err)
	s.ErrorIs(err, domain.ErrNoAvailableReviewers)
}

// TestMultiplePRsForSameUser тестирует множественные PR для одного пользователя
func (s *IntegrationTestSuite) TestMultiplePRsForSameUser() {
	// Создаем команду
	team := domain.Team{
		Name: "multi-pr-team",
		Members: []domain.User{
			{ID: "user-50", Username: "alice", TeamName: "multi-pr-team", IsActive: true},
			{ID: "user-51", Username: "bob", TeamName: "multi-pr-team", IsActive: true},
			{ID: "user-52", Username: "charlie", TeamName: "multi-pr-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем несколько PR
	pr1 := domain.PullRequest{
		ID:       "pr-50",
		Name:     "Feature 1",
		AuthorID: "user-50",
		Status:   domain.PRStatusOpen,
	}
	_, err = s.prService.Create(s.ctx, pr1)
	s.Require().NoError(err)

	pr2 := domain.PullRequest{
		ID:       "pr-51",
		Name:     "Feature 2",
		AuthorID: "user-51",
		Status:   domain.PRStatusOpen,
	}
	_, err = s.prService.Create(s.ctx, pr2)
	s.Require().NoError(err)

	// Проверяем, что user-52 может быть назначен на оба PR
	user52PRs, err := s.prService.GetReviewingPRs(s.ctx, "user-52")
	s.Require().NoError(err)
	s.GreaterOrEqual(len(user52PRs), 1) // Должен иметь хотя бы 1 PR на ревью
}

// TestTeamNotFound тестирует создание команды, которая уже существует
func (s *IntegrationTestSuite) TestTeamAlreadyExists() {
	// Создаем команду
	team := domain.Team{
		Name: "duplicate-team",
		Members: []domain.User{
			{ID: "user-60", Username: "alice", TeamName: "duplicate-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Попытка создать команду с таким же именем
	_, err = s.teamService.Add(s.ctx, team)
	s.Error(err)
	s.ErrorIs(err, domain.ErrTeamAlreadyExists)
}

// TestPRNotFound тестирует попытку мерджа несуществующего PR
func (s *IntegrationTestSuite) TestPRNotFound() {
	_, err := s.prService.Merge(s.ctx, "non-existent-pr")
	s.Error(err)
	s.ErrorIs(err, domain.ErrPRNotFound)
}

// TestUserNotFound тестирует попытку изменения активности несуществующего пользователя
func (s *IntegrationTestSuite) TestUserNotFound() {
	_, err := s.userService.SetIsActive(s.ctx, "non-existent-user", true)
	s.Error(err)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

// TestGetTeam тестирует получение команды
func (s *IntegrationTestSuite) TestGetTeam() {
	// Создаем команду
	team := domain.Team{
		Name: "test-get-team",
		Members: []domain.User{
			{ID: "user-70", Username: "alice", TeamName: "test-get-team", IsActive: true},
			{ID: "user-71", Username: "bob", TeamName: "test-get-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Получаем команду
	retrievedTeam, err := s.teamService.Get(s.ctx, "test-get-team")
	s.Require().NoError(err)
	s.Equal("test-get-team", retrievedTeam.Name)
	s.Len(retrievedTeam.Members, 2)
}

// TestConcurrentPRCreation тестирует одновременное создание PR
func (s *IntegrationTestSuite) TestConcurrentPRCreation() {
	// Создаем команду
	team := domain.Team{
		Name: "concurrent-team",
		Members: []domain.User{
			{ID: "user-80", Username: "alice", TeamName: "concurrent-team", IsActive: true},
			{ID: "user-81", Username: "bob", TeamName: "concurrent-team", IsActive: true},
			{ID: "user-82", Username: "charlie", TeamName: "concurrent-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR одновременно в разных горутинах
	done := make(chan bool)
	for i := 0; i < 3; i++ {
		go func(idx int) {
			pr := domain.PullRequest{
				ID:       fmt.Sprintf("pr-80%d", idx),
				Name:     fmt.Sprintf("Concurrent PR %d", idx),
				AuthorID: "user-80",
				Status:   domain.PRStatusOpen,
			}
			_, err := s.prService.Create(s.ctx, pr)
			s.NoError(err)
			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestIntegrationTestSuite запускает test suite
func TestIntegrationTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests. Set INTEGRATION_TESTS=1 to run them.")
	}
	suite.Run(t, new(IntegrationTestSuite))
}
