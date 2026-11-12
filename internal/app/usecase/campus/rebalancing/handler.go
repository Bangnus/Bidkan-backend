package rebalancing

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{service: service}
}

func (h *Handler) HandleGetOverview(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug is required"})
	}

	overview, err := h.service.GetRebalanceOverview(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": overview})
}
