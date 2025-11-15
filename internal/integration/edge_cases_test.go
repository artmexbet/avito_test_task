package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

// EdgeCasesTestSuite тестирует граничные случаи и edge cases
type EdgeCasesTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *postgres.PostgresContainer
	pool        *pgxpool.Pool
	prService   *service.PullRequestService
	userService *service.UserService
	teamService *service.TeamService
}

// SetupSuite выполняется один раз перед всеми тестами
func (s *EdgeCasesTestSuite) SetupSuite() {
	s.ctx = context.Background()

	migrationsPath, err := filepath.Abs("../../migrations")
	s.Require().NoError(err)

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
				WithStartupTimeout(120*time.Second),
		),
	)
	s.Require().NoError(err)
	s.pgContainer = pgContainer

	connStr, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	pool, err := pgxpool.New(s.ctx, connStr)
	s.Require().NoError(err)
	s.pool = pool

	pg := postgresRepo.New(pool)
	userRepo := repository.NewUserRepository(pg)
	reviewersRepo := repository.NewReviewersRepository(pg)
	prRepo := repository.NewPRRepository(pg)
	teamRepo := repository.NewTeamRepository(pg)

	s.prService = service.NewPullRequestService(prRepo, reviewersRepo, userRepo)
	s.userService = service.NewUserService(userRepo)
	s.teamService = service.NewTeamService(teamRepo, userRepo)
}

// TearDownSuite выполняется один раз после всех тестов
func (s *EdgeCasesTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(s.ctx)
	}
}

// SetupTest выполняется перед каждым тестом
func (s *EdgeCasesTestSuite) SetupTest() {
	_, err := s.pool.Exec(s.ctx, "TRUNCATE TABLE pull_requests_reviewers, pull_requests, users, teams CASCADE")
	s.Require().NoError(err)
}

// TestSingleUserTeam тестирует команду с одним пользователем
func (s *EdgeCasesTestSuite) TestSingleUserTeam() {
	team := domain.Team{
		Name: "solo-team",
		Members: []domain.User{
			{ID: "user-1", Username: "alice", TeamName: "solo-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Попытка создать PR в команде с одним человеком
	pr := domain.PullRequest{
		ID:       "pr-solo",
		Name:     "Solo PR",
		AuthorID: "user-1",
		Status:   domain.PRStatusOpen,
	}

	_, err = s.prService.Create(s.ctx, pr)
	s.Require().Error(err)
}

// TestLargeTeam тестирует команду с большим количеством участников
func (s *EdgeCasesTestSuite) TestLargeTeam() {
	// Создаем команду с 20 участниками
	members := make([]domain.User, 20)
	for i := 0; i < 20; i++ {
		members[i] = domain.User{
			ID:       fmt.Sprintf("user-%d", i),
			Username: fmt.Sprintf("user%d", i),
			TeamName: "large-team",
			IsActive: true,
		}
	}

	team := domain.Team{
		Name:    "large-team",
		Members: members,
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR
	pr := domain.PullRequest{
		ID:       "pr-large",
		Name:     "Large team PR",
		AuthorID: "user-0",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)
	// Должно быть ровно 2 ревьювера
	s.Len(createdPR.Reviewers, 2)

	// Проверяем, что автор не среди ревьюверов
	for _, reviewer := range createdPR.Reviewers {
		s.NotEqual("user-0", reviewer.ID)
	}
}

// TestMultipleReassignments тестирует множественные переназначения
func (s *EdgeCasesTestSuite) TestMultipleReassignments() {
	// Создаем команду с достаточным количеством участников
	members := make([]domain.User, 10)
	for i := 0; i < 10; i++ {
		members[i] = domain.User{
			ID:       fmt.Sprintf("user-%d", i),
			Username: fmt.Sprintf("user%d", i),
			TeamName: "reassign-team",
			IsActive: true,
		}
	}

	team := domain.Team{
		Name:    "reassign-team",
		Members: members,
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR
	pr := domain.PullRequest{
		ID:       "pr-reassign",
		Name:     "Reassignment test",
		AuthorID: "user-0",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)
	s.Len(createdPR.Reviewers, 2)

	// Выполняем несколько переназначений подряд
	currentReviewerID := createdPR.Reviewers[0].ID
	assignedReviewers := map[string]bool{currentReviewerID: true}

	for i := 0; i < 5; i++ {
		_, newReviewerID, err := s.prService.ReassignReviewer(s.ctx, "pr-reassign", currentReviewerID)
		s.Require().NoError(err)

		// Новый ревьювер должен отличаться от старого
		s.NotEqual(currentReviewerID, newReviewerID)

		// Новый ревьювер не должен быть автором
		s.NotEqual("user-0", newReviewerID)

		assignedReviewers[newReviewerID] = true
		currentReviewerID = newReviewerID
	}

	// Проверяем, что было назначено несколько разных ревьюверов
	s.GreaterOrEqual(len(assignedReviewers), 2)
}

// TestInactiveUsersNotAssigned тестирует, что неактивные пользователи не назначаются
func (s *EdgeCasesTestSuite) TestInactiveUsersNotAssigned() {
	team := domain.Team{
		Name: "mixed-team",
		Members: []domain.User{
			{ID: "user-1", Username: "alice", TeamName: "mixed-team", IsActive: true},
			{ID: "user-2", Username: "bob", TeamName: "mixed-team", IsActive: false},
			{ID: "user-3", Username: "charlie", TeamName: "mixed-team", IsActive: false},
			{ID: "user-4", Username: "diana", TeamName: "mixed-team", IsActive: true},
			{ID: "user-5", Username: "eve", TeamName: "mixed-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR от имени активного пользователя
	pr := domain.PullRequest{
		ID:       "pr-mixed",
		Name:     "Mixed team PR",
		AuthorID: "user-1",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)

	// Должно быть 2 ревьювера из активных (user-4 и user-5)
	s.Len(createdPR.Reviewers, 2)

	// Проверяем, что ревьюверы - только активные пользователи
	for _, reviewer := range createdPR.Reviewers {
		s.NotEqual("user-2", reviewer.ID) // bob неактивен
		s.NotEqual("user-3", reviewer.ID) // charlie неактивен
		s.NotEqual("user-1", reviewer.ID) // alice - автор
	}
}

// TestDeactivateUserWithPRs тестирует деактивацию пользователя с PR на ревью
func (s *EdgeCasesTestSuite) TestDeactivateUserWithPRs() {
	team := domain.Team{
		Name: "deactivate-team",
		Members: []domain.User{
			{ID: "user-1", Username: "alice", TeamName: "deactivate-team", IsActive: true},
			{ID: "user-2", Username: "bob", TeamName: "deactivate-team", IsActive: true},
			{ID: "user-3", Username: "charlie", TeamName: "deactivate-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR
	pr := domain.PullRequest{
		ID:       "pr-deactivate",
		Name:     "Deactivate test",
		AuthorID: "user-1",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)
	s.Len(createdPR.Reviewers, 2)

	reviewerID := createdPR.Reviewers[0].ID

	// Проверяем, что у ревьювера есть PR на ревью
	prs, err := s.prService.GetReviewingPRs(s.ctx, reviewerID)
	s.Require().NoError(err)
	s.Len(prs, 1)

	// Деактивируем ревьювера
	_, err = s.userService.SetIsActive(s.ctx, reviewerID, false)
	s.Require().NoError(err)

	// PR все еще должен быть у деактивированного пользователя
	prs, err = s.prService.GetReviewingPRs(s.ctx, reviewerID)
	s.Require().NoError(err)
	s.Len(prs, 1)
}

// TestConcurrentReassignments тестирует конкурентные переназначения
func (s *EdgeCasesTestSuite) TestConcurrentReassignments() {
	// Создаем команду
	members := make([]domain.User, 15)
	for i := 0; i < 15; i++ {
		members[i] = domain.User{
			ID:       fmt.Sprintf("user-%d", i),
			Username: fmt.Sprintf("user%d", i),
			TeamName: "concurrent-team",
			IsActive: true,
		}
	}

	team := domain.Team{
		Name:    "concurrent-team",
		Members: members,
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем PR
	pr := domain.PullRequest{
		ID:       "pr-concurrent",
		Name:     "Concurrent test",
		AuthorID: "user-0",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)

	reviewerID := createdPR.Reviewers[0].ID

	// Пытаемся переназначить одновременно из нескольких горутин
	var wg sync.WaitGroup
	results := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := s.prService.ReassignReviewer(s.ctx, "pr-concurrent", reviewerID)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	// Хотя бы одна операция должна быть успешной
	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		}
	}

	// Может быть несколько успешных операций из-за race condition
	s.GreaterOrEqual(successCount, 1)
}

// TestManyPRsInTeam тестирует создание множества PR в одной команде
func (s *EdgeCasesTestSuite) TestManyPRsInTeam() {
	// Создаем команду
	members := make([]domain.User, 5)
	for i := 0; i < 5; i++ {
		members[i] = domain.User{
			ID:       fmt.Sprintf("user-%d", i),
			Username: fmt.Sprintf("user%d", i),
			TeamName: "many-pr-team",
			IsActive: true,
		}
	}

	team := domain.Team{
		Name:    "many-pr-team",
		Members: members,
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Создаем 50 PR
	for i := 0; i < 50; i++ {
		authorID := fmt.Sprintf("user-%d", i%5)
		pr := domain.PullRequest{
			ID:       fmt.Sprintf("pr-%d", i),
			Name:     fmt.Sprintf("PR %d", i),
			AuthorID: authorID,
			Status:   domain.PRStatusOpen,
		}

		_, err := s.prService.Create(s.ctx, pr)
		s.Require().NoError(err)
	}

	// Проверяем, что у каждого пользователя есть PR на ревью
	for i := 0; i < 5; i++ {
		userID := fmt.Sprintf("user-%d", i)
		prs, err := s.prService.GetReviewingPRs(s.ctx, userID)
		s.Require().NoError(err)
		// Пользователь должен иметь хотя бы один PR на ревью
		s.Greater(len(prs), 0)
	}
}

// TestMergeAndReassign тестирует попытку переназначения после мерджа
func (s *EdgeCasesTestSuite) TestMergeAndReassign() {
	team := domain.Team{
		Name: "merge-reassign-team",
		Members: []domain.User{
			{ID: "user-1", Username: "alice", TeamName: "merge-reassign-team", IsActive: true},
			{ID: "user-2", Username: "bob", TeamName: "merge-reassign-team", IsActive: true},
			{ID: "user-3", Username: "charlie", TeamName: "merge-reassign-team", IsActive: true},
			{ID: "user-4", Username: "diana", TeamName: "merge-reassign-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	pr := domain.PullRequest{
		ID:       "pr-merge-reassign",
		Name:     "Merge reassign test",
		AuthorID: "user-1",
		Status:   domain.PRStatusOpen,
	}

	createdPR, err := s.prService.Create(s.ctx, pr)
	s.Require().NoError(err)

	reviewerID := createdPR.Reviewers[0].ID

	// Мерджим PR
	_, err = s.prService.Merge(s.ctx, "pr-merge-reassign")
	s.Require().NoError(err)

	// Попытка переназначения после мерджа должна провалиться
	_, _, err = s.prService.ReassignReviewer(s.ctx, "pr-merge-reassign", reviewerID)
	s.Error(err)
	// Проверяем конкретную ошибку, если она определена в domain
	// s.ErrorIs(err, domain.ErrCannotReassignMergedPR)
}

// TestUpdateTeamMembers тестирует обновление списка участников команды
func (s *EdgeCasesTestSuite) TestUpdateTeamMembers() {
	// Создаем команду с начальными участниками
	team := domain.Team{
		Name: "update-team",
		Members: []domain.User{
			{ID: "user-1", Username: "alice", TeamName: "update-team", IsActive: true},
			{ID: "user-2", Username: "bob", TeamName: "update-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Получаем команду
	retrievedTeam, err := s.teamService.Get(s.ctx, "update-team")
	s.Require().NoError(err)
	s.Len(retrievedTeam.Members, 2)

	// Попытка добавить команду с тем же именем должна провалиться
	_, err = s.teamService.Add(s.ctx, team)
	s.Error(err)
	s.ErrorIs(err, domain.ErrTeamAlreadyExists)
}

// TestEmptyReviewingList тестирует получение пустого списка PR на ревью
func (s *EdgeCasesTestSuite) TestEmptyReviewingList() {
	team := domain.Team{
		Name: "empty-review-team",
		Members: []domain.User{
			{ID: "user-1", Username: "alice", TeamName: "empty-review-team", IsActive: true},
		},
	}

	_, err := s.teamService.Add(s.ctx, team)
	s.Require().NoError(err)

	// Получаем список PR для пользователя, который ничего не ревьюит
	prs, err := s.prService.GetReviewingPRs(s.ctx, "user-1")
	s.Require().NoError(err)
	s.Len(prs, 0)
}

// TestEdgeCasesTestSuite запускает test suite
func TestEdgeCasesTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping edge cases integration tests. Set INTEGRATION_TESTS=1 to run them.")
	}
	suite.Run(t, new(EdgeCasesTestSuite))
}
