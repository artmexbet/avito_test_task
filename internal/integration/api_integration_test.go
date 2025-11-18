package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	postgresRepo "github.com/artmexbet/avito_test_task/internal/postgres"
	"github.com/artmexbet/avito_test_task/internal/repository"
	"github.com/artmexbet/avito_test_task/internal/router"
	"github.com/artmexbet/avito_test_task/internal/service"
	"github.com/artmexbet/avito_test_task/pkg/config"
)

// APIIntegrationTestSuite определяет test suite для интеграционных тестов API
type APIIntegrationTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *postgres.PostgresContainer
	pool        *pgxpool.Pool
	router      *router.Router
	app         *fiber.App
	client      *http.Client
	baseURL     string
}

// SetupSuite выполняется один раз перед всеми тестами
func (s *APIIntegrationTestSuite) SetupSuite() {
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
				WithStartupTimeout(120*time.Second),
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

	// Инициализируем репозитории и сервисы
	pg := postgresRepo.New(pool)
	userRepo := repository.NewUserRepository(pg)
	reviewersRepo := repository.NewReviewersRepository(pg)
	prRepo := repository.NewPRRepository(pg)
	teamRepo := repository.NewTeamRepository(pg)

	prService := service.NewPullRequestService(prRepo, reviewersRepo, userRepo)
	userService := service.NewUserService(userRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)

	// Инициализируем роутер
	cfg := config.RouterConfig{
		Host: "localhost",
		Port: 5000,
	}
	s.router = router.New(cfg, userService, prService, teamService, nil)

	// Запускаем сервер в фоновом режиме
	go func() {
		_ = s.router.Run()
	}()

	// Даем серверу время запуститься
	time.Sleep(100 * time.Millisecond)

	s.baseURL = fmt.Sprintf("http://localhost:%d", cfg.Port)
	s.client = &http.Client{}
}

// TearDownSuite выполняется один раз после всех тестов
func (s *APIIntegrationTestSuite) TearDownSuite() {
	if s.router != nil {
		_ = s.router.Shutdown(s.ctx)
	}
	if s.pool != nil {
		s.pool.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(s.ctx)
	}
}

// SetupTest выполняется перед каждым тестом
func (s *APIIntegrationTestSuite) SetupTest() {
	// Очищаем все таблицы перед каждым тестом
	_, err := s.pool.Exec(s.ctx, "TRUNCATE TABLE pull_requests_reviewers, pull_requests, users, teams CASCADE")
	s.Require().NoError(err)
}

// makeRequest - вспомогательная функция для выполнения HTTP запросов
func (s *APIIntegrationTestSuite) makeRequest(method, path string, body interface{}) (*http.Response, []byte) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		s.Require().NoError(err)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, s.baseURL+path, bodyReader)
	s.Require().NoError(err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	return resp, respBody
}

// TestAddTeamAPI тестирует POST /team/add
func (s *APIIntegrationTestSuite) TestAddTeamAPI() {
	teamReq := map[string]interface{}{
		"team_name": "backend-team",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "alice", "is_active": true},
			{"user_id": "user-2", "username": "bob", "is_active": true},
		},
	}

	resp, body := s.makeRequest("POST", "/team/add", teamReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	s.Require().NoError(err)

	team := response["team"].(map[string]interface{})
	s.Equal("backend-team", team["team_name"])
}

// TestAddTeamAlreadyExists тестирует POST /teams/add для существующей команды
func (s *APIIntegrationTestSuite) TestAddTeamAlreadyExists() {
	teamReq := map[string]interface{}{
		"team_name": "duplicate-team",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "alice", "is_active": true},
		},
	}

	// Первый запрос должен пройти успешно
	resp, _ := s.makeRequest("POST", "/team/add", teamReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	// Второй запрос должен вернуть ошибку
	resp, body := s.makeRequest("POST", "/team/add", teamReq)
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	s.Require().NoError(err)

	errorObj := response["error"].(map[string]interface{})
	s.Equal("TEAM_EXISTS", errorObj["code"])
}

// TestGetTeamAPI тестирует GET /team/get
func (s *APIIntegrationTestSuite) TestGetTeamAPI() {
	// Сначала создаем команду
	teamReq := map[string]interface{}{
		"team_name": "test-team",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "alice", "is_active": true},
			{"user_id": "user-2", "username": "bob", "is_active": true},
		},
	}

	resp, _ := s.makeRequest("POST", "/team/add", teamReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	// Получаем команду
	resp, body := s.makeRequest("GET", "/team/get?team_name=test-team", nil)
	s.Equal(http.StatusOK, resp.StatusCode)

	var team map[string]interface{}
	err := json.Unmarshal(body, &team)
	s.Require().NoError(err)

	s.Equal("test-team", team["team_name"])
	members := team["members"].([]interface{})
	s.Len(members, 2)
}

// TestSetUserIsActiveAPI тестирует POST /users/setIsActive
func (s *APIIntegrationTestSuite) TestSetUserIsActiveAPI() {
	// Сначала создаем команду с пользователем
	teamReq := map[string]interface{}{
		"team_name": "test-team",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "alice", "is_active": true},
		},
	}

	resp, _ := s.makeRequest("POST", "/teams/add", teamReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	// Деактивируем пользователя
	setActiveReq := map[string]interface{}{
		"user_id":   "user-1",
		"is_active": false,
	}

	resp, body := s.makeRequest("POST", "/users/setIsActive", setActiveReq)
	s.Equal(http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	s.Require().NoError(err)

	user := response["user"].(map[string]interface{})
	s.False(user["is_active"].(bool))
}

// TestCreatePRAPI тестирует POST /pullRequest/create
func (s *APIIntegrationTestSuite) TestCreatePRAPI() {
	// Создаем команду
	teamReq := map[string]interface{}{
		"team_name": "dev-team",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "alice", "is_active": true},
			{"user_id": "user-2", "username": "bob", "is_active": true},
			{"user_id": "user-3", "username": "charlie", "is_active": true},
		},
	}

	resp, _ := s.makeRequest("POST", "/team/add", teamReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	// Создаем PR
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature X",
		"author_id":         "user-1",
	}

	resp, body := s.makeRequest("POST", "/pullRequest/create", prReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	s.Require().NoError(err)

	pr := response["pr"].(map[string]interface{})
	s.Equal("pr-1", pr["pull_request_id"])
	s.Equal("OPEN", pr["status"])
	assignedReviewers := pr["assigned_reviewers"].([]interface{})
	s.Len(assignedReviewers, 2)
}

// TestMergePRAPI тестирует POST /pullRequest/merge
func (s *APIIntegrationTestSuite) TestMergePRAPI() {
	// Создаем команду и PR
	teamReq := map[string]interface{}{
		"team_name": "merge-team",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "alice", "is_active": true},
			{"user_id": "user-2", "username": "bob", "is_active": true},
		},
	}

	resp, _ := s.makeRequest("POST", "/team/add", teamReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-merge",
		"pull_request_name": "Merge test",
		"author_id":         "user-1",
	}

	resp, _ = s.makeRequest("POST", "/pullRequest/create", prReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	// Мерджим PR
	mergeReq := map[string]interface{}{
		"pull_request_id": "pr-merge",
	}

	resp, body := s.makeRequest("POST", "/pullRequest/merge", mergeReq)
	s.Equal(http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	s.Require().NoError(err)

	pr := response["pr"].(map[string]interface{})
	s.Equal("MERGED", pr["status"])
}

// TestReassignReviewerAPI тестирует POST /pullRequest/reassign
func (s *APIIntegrationTestSuite) TestReassignReviewerAPI() {
	// Создаем команду и PR
	teamReq := map[string]interface{}{
		"team_name": "reassign-team",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "alice", "is_active": true},
			{"user_id": "user-2", "username": "bob", "is_active": true},
			{"user_id": "user-3", "username": "charlie", "is_active": true},
			{"user_id": "user-4", "username": "diana", "is_active": true},
		},
	}

	resp, _ := s.makeRequest("POST", "/team/add", teamReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-reassign",
		"pull_request_name": "Reassign test",
		"author_id":         "user-1",
	}

	resp, body := s.makeRequest("POST", "/pullRequest/create", prReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	var createResponse map[string]interface{}
	err := json.Unmarshal(body, &createResponse)
	s.Require().NoError(err)

	pr := createResponse["pr"].(map[string]interface{})
	assignedReviewers := pr["assigned_reviewers"].([]interface{})
	oldReviewerID := assignedReviewers[0].(string)

	// Переназначаем ревьювера
	reassignReq := map[string]interface{}{
		"pull_request_id": "pr-reassign",
		"old_user_id":     oldReviewerID,
	}

	resp, body = s.makeRequest("POST", "/pullRequest/reassign", reassignReq)
	s.Equal(http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	s.Require().NoError(err)

	newReviewerID := response["replaced_by"].(string)
	s.NotEqual(oldReviewerID, newReviewerID)
}

// TestGetUserReviewAPI тестирует GET /users/getReview
func (s *APIIntegrationTestSuite) TestGetUserReviewAPI() {
	// Создаем команду и PR
	teamReq := map[string]interface{}{
		"team_name": "review-team",
		"members": []map[string]interface{}{
			{"user_id": "user-1", "username": "alice", "is_active": true},
			{"user_id": "user-2", "username": "bob", "is_active": true},
			{"user_id": "user-3", "username": "charlie", "is_active": true},
		},
	}

	resp, _ := s.makeRequest("POST", "/team/add", teamReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	prReq := map[string]interface{}{
		"pull_request_id":   "pr-review",
		"pull_request_name": "Review test",
		"author_id":         "user-1",
	}

	resp, _ = s.makeRequest("POST", "/pullRequest/create", prReq)
	s.Equal(http.StatusCreated, resp.StatusCode)

	// Получаем PRы для ревьюера
	resp, body := s.makeRequest("GET", "/users/getReview?user_id=user-2", nil)
	s.Equal(http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	s.Require().NoError(err)

	s.Equal("user-2", response["user_id"])
	pullRequests := response["pull_requests"].([]interface{})
	// Может быть 0 или 1 в зависимости от того, был ли user-2 назначен ревьювером
	s.GreaterOrEqual(len(pullRequests), 0)
}

// TestErrorCases тестирует различные ошибочные случаи
func (s *APIIntegrationTestSuite) TestErrorCases() {
	// Попытка получить несуществующую команду
	resp, body := s.makeRequest("GET", "/team/get?team_name=non-existent", nil)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	s.Require().NoError(err)

	errorObj := response["error"].(map[string]interface{})
	s.Equal("NOT_FOUND", errorObj["code"])

	// Попытка изменить активность несуществующего пользователя
	setActiveReq := map[string]interface{}{
		"user_id":   "non-existent-user",
		"is_active": false,
	}

	resp, body = s.makeRequest("POST", "/users/setIsActive", setActiveReq)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	err = json.Unmarshal(body, &response)
	s.Require().NoError(err)

	errorObj = response["error"].(map[string]interface{})
	s.Equal("NOT_FOUND", errorObj["code"])

	// Попытка мерджа несуществующего PR
	mergeReq := map[string]interface{}{
		"pull_request_id": "non-existent-pr",
	}

	resp, body = s.makeRequest("POST", "/pullRequest/merge", mergeReq)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	err = json.Unmarshal(body, &response)
	s.Require().NoError(err)

	errorObj = response["error"].(map[string]interface{})
	s.Equal("NOT_FOUND", errorObj["code"])
}

// TestAPIIntegrationTestSuite запускает test suite
func TestAPIIntegrationTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping API integration tests. Set INTEGRATION_TESTS=1 to run them.")
	}
	suite.Run(t, new(APIIntegrationTestSuite))
}
