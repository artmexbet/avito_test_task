package router

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/artmexbet/avito_test_task/internal/domain"
)

func (r *Router) createPullRequest(ctx *fiber.Ctx) error {
	uCtx := ctx.UserContext()

	var req createPRRequest
	if err := ctx.BodyParser(&req); err != nil {
		slog.ErrorContext(uCtx, "failed to parse create PR request", "error", err)
		return fiber.ErrBadRequest
	}
	if err := r.validator.StructCtx(uCtx, req); err != nil {
		slog.WarnContext(uCtx, "validation failed for create PR request", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(errorBadRequest)
	}

	pr, err := r.pullRequestService.Create(uCtx, req.ToDomain())
	if err != nil {
		if errors.Is(err, domain.ErrPRAlreadyExists) {
			slog.WarnContext(uCtx, "pull request already exists", "pr_id", req.PullRequestID)
			return ctx.Status(fiber.StatusConflict).JSON(
				newErrorResponse(
					fmt.Sprintf("%s already exists", req.PullRequestID),
					errorCodePRExists,
				),
			)
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			slog.WarnContext(uCtx, "author not found on PR create", "author_id", req.AuthorID)
			return ctx.Status(fiber.StatusNotFound).JSON(errorResponseNotFound)
		}
		slog.ErrorContext(uCtx, "failed to create PR", "error", err)
		return fiber.ErrInternalServerError
	}

	resp := fromDomainPR(pr)
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"pr": resp})
}

func (r *Router) mergePullRequest(ctx *fiber.Ctx) error {
	uCtx := ctx.UserContext()

	var req mergePRRequest
	if err := ctx.BodyParser(&req); err != nil {
		slog.ErrorContext(uCtx, "failed to parse merge PR request", "error", err)
		return fiber.ErrBadRequest
	}
	if err := r.validator.StructCtx(uCtx, req); err != nil {
		slog.WarnContext(uCtx, "validation failed for merge PR request", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(errorBadRequest)
	}

	pr, err := r.pullRequestService.Merge(uCtx, req.PullRequestID)
	if err != nil {
		if errors.Is(err, domain.ErrPRNotFound) {
			slog.WarnContext(uCtx, "pull request not found on merge", "pr_id", req.PullRequestID)
			return ctx.Status(fiber.StatusNotFound).JSON(errorResponseNotFound)
		}
		if errors.Is(err, domain.ErrPRAlreadyMerged) {
			slog.WarnContext(uCtx, "pull request already merged", "pr_id", req.PullRequestID)
			return ctx.Status(fiber.StatusConflict).JSON(
				newErrorResponse("pull request already merged", errorCodePRMerged),
			)
		}
		slog.ErrorContext(uCtx, "failed to merge PR", "error", err)
		return fiber.ErrInternalServerError
	}

	resp := fromDomainPR(pr)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"pr": resp})
}

func (r *Router) reassignReviewer(ctx *fiber.Ctx) error {
	uCtx := ctx.UserContext()

	var req reassignReviewerRequest
	if err := ctx.BodyParser(&req); err != nil {
		slog.ErrorContext(uCtx, "failed to parse reassign request", "error", err)
		return fiber.ErrBadRequest
	}
	if err := r.validator.StructCtx(uCtx, req); err != nil {
		slog.WarnContext(uCtx, "validation failed for reassign request", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(errorBadRequest)
	}

	pr, newID, err := r.pullRequestService.ReassignReviewer(uCtx, req.PullRequestID, req.OldUserID)
	if err != nil {
		if errors.Is(err, domain.ErrPRNotFound) || errors.Is(err, domain.ErrUserNotFound) {
			slog.WarnContext(
				uCtx,
				"pr or user not found on reassign",
				"pr_id", req.PullRequestID,
				"old_user_id", req.OldUserID,
			)
			return ctx.Status(fiber.StatusNotFound).JSON(errorResponseNotFound)
		}
		if errors.Is(err, domain.ErrReviewerNotAssigned) {
			slog.WarnContext(uCtx, "reviewer not assigned to PR", "pr_id", req.PullRequestID, "old_user_id", req.OldUserID)
			return ctx.Status(fiber.StatusConflict).JSON(
				newErrorResponse("reviewer is not assigned to this PR", errorCodeNotAssigned),
			)
		}
		if errors.Is(err, domain.ErrNoAvailableReviewers) {
			slog.WarnContext(
				uCtx,
				"no active replacement candidate in team",
				"pr_id", req.PullRequestID,
				"old_user_id", req.OldUserID,
			)
			return ctx.Status(fiber.StatusConflict).JSON(
				newErrorResponse("no active replacement candidate in team", errorCodeNoCandidate),
			)
		}
		slog.ErrorContext(uCtx, "failed to reassign reviewer", "error", err)
		return fiber.ErrInternalServerError
	}

	resp := reassignReviewerResponse{PR: fromDomainPR(*pr), ReplacedBy: newID}
	return ctx.Status(fiber.StatusOK).JSON(resp)
}
