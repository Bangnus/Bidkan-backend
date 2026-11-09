package revenue_manage

import (
	"fmt"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UpdateRatesRequest struct {
	UnlockFee     float64 `json:"unlock_fee"`
	PerMinuteRate float64 `json:"per_minute_rate"`
}

type Handler struct {
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
}

func NewHandler(tenantRepo domainRepo.TenantRepository, campusRepo domainRepo.CampusOpsRepository) Handler {
	return Handler{tenantRepo: tenantRepo, campusRepo: campusRepo}
}

func (h *Handler) HandleGetRevenue(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	payouts, err := h.campusRepo.ListPayoutsByTenant(c.Context(), tenant.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var totalGross, totalFee, totalNet float64
	for _, p := range payouts {
		totalGross += p.GrossAmount
		totalFee += p.PlatformFeeAmount
		totalNet += p.NetPayoutAmount
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"tenant":       tenant,
			"total_gross":  totalGross,
			"total_gp_fee": totalFee,
			"total_net":    totalNet,
			"payouts":      payouts,
		},
	})
}

func (h *Handler) HandleUpdateRates(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	var req UpdateRatesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.tenantRepo.UpdateRates(c.Context(), tenant.ID, req.UnlockFee, req.PerMinuteRate, tenant.PlatformFeePercent); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "campus pricing rates updated"})
}

func (h *Handler) HandleStripeOnboard(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	acctID := fmt.Sprintf("acct_%s_%s", tenant.Slug, uuid.NewString()[:6])
	if tenant.StripeAccountID != nil && *tenant.StripeAccountID != "" {
		acctID = *tenant.StripeAccountID
	}
	if err := h.campusRepo.CompleteStripeOnboarding(c.Context(), tenant.ID, acctID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":           "Stripe Connect onboarding completed",
		"stripe_account_id": acctID,
	})
}
