package router

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/artmexbet/avito_test_task/pkg/config"
)

type iUserService interface {
	// Define user service methods here
}

type iPullRequestService interface {
	// Define pull request service methods here
}

type iTeamService interface {
}

type Config struct {
	Host string `yaml:"host" env:"HOST"`
	Port int    `yaml:"port" env:"PORT"`
}

type Router struct {
	config config.RouterConfig

	router             *fiber.App
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

	router := &Router{
		config:             config,
		router:             app,
		userService:        userService,
		pullRequestService: pullRequestService,
		teamService:        teamService,
	}
	router.initRoutes()

	return router
}

func (r *Router) initRoutes() {

}

func (r *Router) Run() error {
	addr := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)

	return r.router.Listen(addr)
}

func (r *Router) Shutdown(ctx context.Context) error {
	return r.router.ShutdownWithContext(ctx)
}
