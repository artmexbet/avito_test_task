package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	_recover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	slogfiber "github.com/samber/slog-fiber"

	"github.com/artmexbet/avito_test_task/internal/domain"
	stats_retriever "github.com/artmexbet/avito_test_task/internal/stats-retriever"
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

type iStatsRetriever interface {
	RetrieveStats(ctx context.Context) ([]stats_retriever.Stats, error)
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
	statsRetriever     iStatsRetriever
}

func New(
	config config.RouterConfig,
	userService iUserService,
	pullRequestService iPullRequestService,
	teamService iTeamService,
	statsRetriever iStatsRetriever,
) *Router {
	app := fiber.New()

	router := &Router{
		config:             config,
		router:             app,
		userService:        userService,
		pullRequestService: pullRequestService,
		teamService:        teamService,
		statsRetriever:     statsRetriever,
		validator:          validator.New(validator.WithRequiredStructEnabled()),
	}
	router.initMiddlewares()
	router.initRoutes()

	return router
}

func (r *Router) initMiddlewares() {
	r.router.Use(slogfiber.New(slog.Default()))
	r.router.Use(_recover.New())
	r.router.Use(healthcheck.New())
	r.router.Use(requestid.New())
	r.router.Use(
		swaggerui.New(
			swaggerui.Config{ //nolint:exhaustruct
				BasePath: "/",
				Path:     "docs",
				FilePath: "./docs/openapi.yml",
			},
		),
	)
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

	if r.statsRetriever == nil {
		return
	}

	r.router.Get("/stats/get", func(c *fiber.Ctx) error {
		s, err := r.statsRetriever.RetrieveStats(c.UserContext())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to retrieve stats: %v", err))
		}
		return c.JSON(s)
	})
}

func (r *Router) Run() error {
	addr := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)

	return r.router.Listen(addr)
}

func (r *Router) Shutdown(ctx context.Context) error {
	return r.router.ShutdownWithContext(ctx)
}
