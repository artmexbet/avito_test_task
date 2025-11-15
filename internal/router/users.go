package router

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

func (r *Router) setUserIsActive(ctx *fiber.Ctx) error {
	uCtx := ctx.UserContext()

	var req setUserIsActiveRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		slog.ErrorContext(uCtx, "failed to parse set user is active request", "error", err)
		return fiber.ErrBadRequest
	}

	if r.validator.StructCtx(uCtx, &req) != nil {
		slog.WarnContext(uCtx, "validation failed for set user is active request", "request", req)
		return ctx.Status(fiber.StatusBadRequest).JSON(errorBadRequest)
	}

	user, err := r.userService.SetIsActive(uCtx, req.UserID, req.IsActive)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		slog.ErrorContext(uCtx, "failed to set user is active",
			"error", err,
			"user_id", req.UserID,
			"is_active", req.IsActive)
		// Here you can add more specific error handling based on the type of error
		return fiber.ErrInternalServerError
	} else if errors.Is(err, domain.ErrUserNotFound) {
		slog.WarnContext(uCtx, "user not found when setting is active",
			"user_id", req.UserID)
		return ctx.Status(fiber.StatusNotFound).JSON(errorResponseNotFound)
	}

	return ctx.JSON(fiber.Map{"user": fromDomainUser(user)})
}

func (r *Router) getUserReview(ctx *fiber.Ctx) error {
	uCtx := ctx.UserContext()

	userID := ctx.Query("user_id")
	if userID == "" {
		slog.WarnContext(uCtx, "user_id query param is required")
		return ctx.Status(fiber.StatusBadRequest).JSON(errorBadRequest)
	}

	prs, err := r.pullRequestService.GetReviewingPRs(uCtx, userID)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		slog.ErrorContext(uCtx, "failed to get reviewing PRs", "error", err, "user_id", userID)
		return fiber.ErrInternalServerError
	} else if errors.Is(err, domain.ErrUserNotFound) {
		slog.WarnContext(uCtx, "user not found when getting review PRs", "user_id", userID)
		return ctx.Status(fiber.StatusNotFound).JSON(errorResponseNotFound)
	}

	resp := reviewPRsResponse{
		UserID:       userID,
		PullRequests: make([]pullRequestShortResponse, 0, len(prs)),
	}

	for _, pr := range prs {
		resp.PullRequests = append(resp.PullRequests, pullRequestShortResponse{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(resp)
}
