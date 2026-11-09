package bike_manage

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
)

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type Handler struct {
	campusRepo domainRepo.CampusOpsRepository
}

func NewHandler(campusRepo domainRepo.CampusOpsRepository) Handler {
	return Handler{campusRepo: campusRepo}
}

func (h *Handler) HandleUpdateStatus(c *fiber.Ctx) error {
	bikeID := c.Params("id")
	if bikeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bike id is required"})
	}
	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil || req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status is required"})
	}

	if err := h.campusRepo.UpdateBikeStatus(c.Context(), bikeID, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "bike status updated"})
}
