package bike_manage

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateBikeRequest struct {
	BikeID     string  `json:"bike_id"`
	HardwareID string  `json:"hardware_id"`
	TenantSlug *string `json:"tenant_slug"`
}

type Handler struct {
	superRepo  domainRepo.SuperAdminRepository
	tenantRepo domainRepo.TenantRepository
}

func NewHandler(superRepo domainRepo.SuperAdminRepository, tenantRepo domainRepo.TenantRepository) Handler {
	return Handler{superRepo: superRepo, tenantRepo: tenantRepo}
}

func (h *Handler) HandleListAll(c *fiber.Ctx) error {
	bikes, err := h.superRepo.ListAllBikesWithTenant(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": bikes})
}

func (h *Handler) HandleCreate(c *fiber.Ctx) error {
	var req CreateBikeRequest
	if err := c.BodyParser(&req); err != nil || req.BikeID == "" || req.HardwareID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bike_id and hardware_id are required"})
	}

	var tenantID *uuid.UUID
	if req.TenantSlug != nil && *req.TenantSlug != "" && *req.TenantSlug != "none" {
		t, err := h.tenantRepo.GetBySlug(c.Context(), *req.TenantSlug)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "target campus tenant not found"})
		}
		tenantID = &t.ID
	}

	if err := h.superRepo.CreateBikeWithTenant(c.Context(), req.BikeID, req.HardwareID, tenantID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if tenantID != nil {
		_, _ = h.tenantRepo.RecalculateTieredGP(c.Context(), *tenantID)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "bike provisioned successfully"})
}
