package field_ops

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{service: service}
}

func (h *Handler) HandleGetQueue(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "slug is required",
		})
	}

	queue, err := h.service.GetQueue(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": queue,
	})
}

func (h *Handler) HandleSwapBattery(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var req SwapBatteryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var staffUserID *uuid.UUID
	if uidStr, ok := c.Locals("user_id").(string); ok {
		if uid, err := uuid.Parse(uidStr); err == nil {
			staffUserID = &uid
		}
	}

	if err := h.service.SwapBattery(c.Context(), slug, req, staffUserID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Battery swapped successfully to 100%",
	})
}
