package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artmexbet/avito_test_task/internal/postgres"
	"github.com/artmexbet/avito_test_task/internal/repository"
	"github.com/artmexbet/avito_test_task/internal/router"
	"github.com/artmexbet/avito_test_task/internal/service"
	statsRetriever "github.com/artmexbet/avito_test_task/internal/stats-retriever"
	"github.com/artmexbet/avito_test_task/pkg/config"
	"github.com/artmexbet/avito_test_task/pkg/logger"
)

func main() {
	// Можно было весь этот код выкинуть в отдельную структуру App, инициализировать её и запускать методы Start и Stop
	// Но для простоты задания оставил всё в main.go
	cfg := config.MustParseConfig(config.SourceEnv)

	slog.SetDefault(logger.NewLogger(logger.EnvDevelopment))

	ctx := context.Background()
	slog.InfoContext(ctx, "reading configuration completed", "config", cfg)

	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		panic(err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		panic(err)
	}
	slog.InfoContext(ctx, "connected to postgres database")

	pg := postgres.New(pool)

	userRepository := repository.NewUserRepository(pg)
	reviewersRepository := repository.NewReviewersRepository(pg)
	pullRequestRepository := repository.NewPRRepository(pg)
	teamRepository := repository.NewTeamRepository(pg)

	statsRepository := repository.NewStatsRepository(pg)
	slog.InfoContext(ctx, "repositories initialized")

	prService := service.NewPullRequestService(pullRequestRepository, reviewersRepository, userRepository)
	userService := service.NewUserService(userRepository)
	teamService := service.NewTeamService(teamRepository, userRepository)

	statsService := statsRetriever.NewStatsRetriever(statsRepository)

	_router := router.New(cfg.Router, userService, prService, teamService, statsService)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := _router.Run()
		if err != nil {
			panic(err)
		}
	}()

	<-quit // wait for shutdown signal

	slog.InfoContext(ctx, "shutting down server...")
	ctx, cancel := context.WithTimeout(ctx, cfg.Router.ShutdownTimeout)
	defer cancel()
	err = _router.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	pool.Close()
	slog.InfoContext(ctx, "server gracefully stopped")
}
