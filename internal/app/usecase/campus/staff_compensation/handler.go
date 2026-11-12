package staff_compensation

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

func (h *Handler) HandleGetConfig(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug is required"})
	}

	cfg, err := h.service.GetConfig(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": cfg})
}

func (h *Handler) HandleUpdateConfig(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug is required"})
	}

	var req UpdateConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	cfg, err := h.service.UpdateConfig(c.Context(), slug, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "อัปเดตโมเดลค่าตอบแทนช่างเรียบร้อยแล้ว",
		"data":    cfg,
	})
}

func (h *Handler) HandleListEarnings(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug is required"})
	}

	list, err := h.service.ListStaffEarnings(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": list})
}

func (h *Handler) HandleGetMyEarnings(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug is required"})
	}

	var staffUserID uuid.UUID
	if uidStr, ok := c.Locals("user_id").(string); ok {
		if uid, err := uuid.Parse(uidStr); err == nil {
			staffUserID = uid
		}
	}
	if staffUserID == uuid.Nil {
		// Fallback for unauthenticated dev inspection
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": fiber.Map{
				"today_swaps": 0, "today_repairs": 0, "today_earnings": 0.0,
				"month_swaps": 0, "month_repairs": 0, "month_earnings": 0.0,
				"pay_model": "per_task", "battery_swap_fee": 20.0, "repair_task_fee": 50.0,
			},
		})
	}

	summary, err := h.service.GetMyEarnings(c.Context(), slug, staffUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": summary})
}
