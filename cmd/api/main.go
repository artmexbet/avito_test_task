package main

import (
	"avito_test_task/internal/postgres"
	"avito_test_task/internal/repository"
	"avito_test_task/internal/router"
	"avito_test_task/internal/service"
	"avito_test_task/pkg/config"

	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Можно было весь этот код выкинуть в отдельную структуру App, инициализировать её и запускать методы Start и Stop
	// Но для простоты задания оставил всё в main.go
	cfg := config.MustParseConfig(config.SourceEnv)

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		panic(err)
	}

	pg := postgres.New(pool)

	userRepository := repository.NewUserRepository(pg)
	reviewersRepository := repository.NewReviewersRepository(pg)
	pullRequestRepository := repository.NewPRRepository(pg)

	prService := service.NewPullRequestService(pullRequestRepository, reviewersRepository)
	userService := service.NewUserService(userRepository)

	_router := router.New(cfg.Router, userService, prService)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := _router.Run()
		if err != nil {
			panic(err)
		}
	}()

	<-quit // wait for shutdown signal

	err = _router.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	pool.Close()
}
