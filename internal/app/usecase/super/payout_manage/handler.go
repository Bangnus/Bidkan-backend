package payout_manage

import (
	"errors"

	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svc Service
}

func NewHandler(
	superRepo domainRepo.SuperAdminRepository,
	tenantRepo domainRepo.TenantRepository,
	campusRepo domainRepo.CampusOpsRepository,
	stripeClient infraStripe.Client,
) Handler {
	return Handler{svc: NewService(superRepo, tenantRepo, campusRepo, stripeClient)}
}

func NewHandlerWithService(svc Service) Handler {
	return Handler{svc: svc}
}

func (h *Handler) Service() Service {
	return h.svc
}

func (h *Handler) HandleListAll(c *fiber.Ctx) error {
	payouts, unsettled, configured, autoEnabled, err := h.svc.ListOverview(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":                payouts,
		"unsettled":           unsettled,
		"stripe_configured":   configured,
		"auto_payout_enabled": autoEnabled,
	})
}

func (h *Handler) HandleTriggerPayout(c *fiber.Ctx) error {
	var req TriggerPayoutRequest
	if err := c.BodyParser(&req); err != nil || req.TenantSlug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tenant_slug is required"})
	}
	payout, err := h.svc.TriggerPayout(c.Context(), req)
	if err != nil {
		return writeServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Stripe Payout completed",
		"data":    payout,
	})
}

func (h *Handler) HandleRetryPayout(c *fiber.Ctx) error {
	payoutID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payout id"})
	}
	payout, err := h.svc.RetryPayout(c.Context(), payoutID)
	if err != nil {
		return writeServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Stripe Payout retry completed",
		"data":    payout,
	})
}

func (h *Handler) HandleRunAuto(c *fiber.Ctx) error {
	res, err := h.svc.RunAutoPayouts(c.Context())
	if err != nil {
		return writeServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Auto-Payout batch executed",
		"data":    res,
	})
}

func writeServiceError(c *fiber.Ctx, err error) error {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		payload := fiber.Map{"error": svcErr.Msg}
		if svcErr.Code != "" {
			payload["code"] = svcErr.Code
		}
		if svcErr.Payout != nil {
			payload["data"] = svcErr.Payout
		}
		return c.Status(svcErr.Status).JSON(payload)
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}
