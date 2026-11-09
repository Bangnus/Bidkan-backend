package report_manage

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
}

func NewHandler(tenantRepo domainRepo.TenantRepository, campusRepo domainRepo.CampusOpsRepository) Handler {
	return Handler{tenantRepo: tenantRepo, campusRepo: campusRepo}
}

func (h *Handler) HandleList(c *fiber.Ctx) error {
	tenant, err := h.tenantRepo.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "campus tenant not found"})
	}
	reports, err := h.campusRepo.ListReportsByTenant(c.Context(), tenant.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": reports})
}

func (h *Handler) HandleResolve(c *fiber.Ctx) error {
	reportID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid report id"})
	}
	var resolvedBy *uuid.UUID
	if uidStr, ok := c.Locals("user_id").(string); ok {
		if uid, err := uuid.Parse(uidStr); err == nil {
			resolvedBy = &uid
		}
	}

	if err := h.campusRepo.ResolveReport(c.Context(), reportID, resolvedBy); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "maintenance report resolved"})
}
