package ride_list

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
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
	rides, err := h.campusRepo.ListRidesByTenant(c.Context(), tenant.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": rides})
}
