package router

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/artmexbet/avito_test_task/internal/domain"
	"github.com/artmexbet/avito_test_task/pkg/config"
)

type iUserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
}

type iPullRequestService interface {
	GetReviewingPRs(ctx context.Context, userID string) ([]domain.PullRequest, error)
	Create(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error)
	Merge(ctx context.Context, prID string) (domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error)
}

type iTeamService interface {
	Add(ctx context.Context, team domain.Team) (domain.Team, error)
	Get(ctx context.Context, teamName string) (domain.Team, error)
}

type Config struct {
	Host string `yaml:"host" env:"HOST"`
	Port int    `yaml:"port" env:"PORT"`
}

type Router struct {
	config config.RouterConfig

	router    *fiber.App
	validator *validator.Validate

	userService        iUserService
	pullRequestService iPullRequestService
	teamService        iTeamService
}

func New(
	config config.RouterConfig,
	userService iUserService,
	pullRequestService iPullRequestService,
	teamService iTeamService,
) *Router {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	router := &Router{
		config:             config,
		router:             app,
		userService:        userService,
		pullRequestService: pullRequestService,
		teamService:        teamService,
		validator:          validator.New(validator.WithRequiredStructEnabled()),
	}
	router.initRoutes()

	return router
}

func (r *Router) initRoutes() {
	teams := r.router.Group("/teams")
	teams.Post("/add", r.addTeam)
	teams.Get("/get", r.getTeam)

	users := r.router.Group("/users")
	users.Post("/setIsActive", r.setUserIsActive)
	users.Get("/getReview", r.getUserReview)

	prs := r.router.Group("/pullRequest")
	prs.Post("/create", r.createPullRequest)
	prs.Post("/merge", r.mergePullRequest)
	prs.Post("/reassign", r.reassignReviewer)
}

func (r *Router) Run() error {
	addr := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)

	return r.router.Listen(addr)
}

func (r *Router) Shutdown(ctx context.Context) error {
	return r.router.ShutdownWithContext(ctx)
}
