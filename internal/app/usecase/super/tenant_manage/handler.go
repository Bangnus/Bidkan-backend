package tenant_manage

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StatusRequest struct {
	Status string `json:"status"`
}

type RatesRequest struct {
	UnlockFee          float64 `json:"unlock_fee"`
	PerMinuteRate      float64 `json:"per_minute_rate"`
	PlatformFeePercent float64 `json:"platform_fee_percent"`
}

type Handler struct {
	tenantRepo domainRepo.TenantRepository
	superRepo  domainRepo.SuperAdminRepository
}

func NewHandler(tenantRepo domainRepo.TenantRepository, superRepo domainRepo.SuperAdminRepository) Handler {
	return Handler{tenantRepo: tenantRepo, superRepo: superRepo}
}

func (h *Handler) HandleUpdateStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tenant id"})
	}
	var req StatusRequest
	if err := c.BodyParser(&req); err != nil || (req.Status != "active" && req.Status != "suspended") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be active or suspended"})
	}
	if err := h.superRepo.UpdateTenantStatus(c.Context(), id, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "tenant status updated"})
}

func (h *Handler) HandleUpdateRates(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tenant id"})
	}
	var req RatesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.tenantRepo.UpdateRates(c.Context(), id, req.UnlockFee, req.PerMinuteRate, req.PlatformFeePercent); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "tenant rates updated"})
}
