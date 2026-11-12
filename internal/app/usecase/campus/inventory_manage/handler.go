package inventory_manage

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetInventory(c *fiber.Ctx) error {
	slug := c.Params("slug")
	category := c.Query("category")
	summary, err := h.service.GetInventory(c.Context(), slug, category)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": summary})
}

func (h *Handler) AdjustStock(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var req AdjustStockReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	actorName, _ := c.Locals("user_name").(string)
	if actorName != "" {
		req.ActorName = actorName
	}

	part, err := h.service.AdjustStock(c.Context(), slug, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": part})
}

func (h *Handler) CreatePart(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var req CreatePartReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	part, err := h.service.CreatePart(c.Context(), slug, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": part})
}

func (h *Handler) ListLogs(c *fiber.Ctx) error {
	slug := c.Params("slug")
	partIDStr := c.Query("part_id")
	var partID *uuid.UUID
	if p, err := uuid.Parse(partIDStr); err == nil {
		partID = &p
	}

	logs, err := h.service.ListUsageLogs(c.Context(), slug, partID, 50)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": logs})
}
