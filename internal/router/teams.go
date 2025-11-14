package router

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

func (r *Router) addTeam(ctx *fiber.Ctx) error {
	uCtx := ctx.UserContext()

	var req addTeamRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		slog.ErrorContext(uCtx, "failed to unmarshal add team request", "error", err)
		return err
	}

	if err := r.validator.StructCtx(uCtx, req); err != nil {
		slog.ErrorContext(uCtx, "validation failed for add team request", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(errorBadRequest)
	}

	team, err := r.teamService.Add(uCtx, req.ToDomain())
	if err != nil && !errors.Is(err, domain.ErrTeamAlreadyExists) {
		slog.ErrorContext(uCtx, "failed to add team", "error", err)
		return err
	} else if errors.Is(err, domain.ErrTeamAlreadyExists) {
		slog.ErrorContext(uCtx, "team already exists", "team_name", req.TeamName)
		return ctx.Status(fiber.StatusBadRequest).
			JSON(
				newErrorResponse(
					fmt.Sprintf("%s already exists", req.TeamName),
					errorCodeTeamExist,
				),
			)
	}

	resp := fromDomainTeam(team)

	// Не стал делать отдельную структуру ответа, нет нужды здесь
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"team": resp})
}

func (r *Router) getTeam(ctx *fiber.Ctx) error {
	uCtx := ctx.UserContext()
	teamName := ctx.Query("team_name")
	if teamName == "" {
		slog.ErrorContext(uCtx, "team_name query param is required")
		return fiber.ErrBadRequest
	}

	team, err := r.teamService.Get(uCtx, teamName)
	if err != nil {
		slog.ErrorContext(uCtx, "failed to get team", "error", err)
		return ctx.Status(fiber.StatusNotFound).JSON(errorResponseNotFound)
	}

	resp := fromDomainTeam(team)

	return ctx.Status(fiber.StatusOK).JSON(resp)
}
