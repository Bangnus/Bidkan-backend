package rider_manage

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

func (h *Handler) HandleListRiders(c *fiber.Ctx) error {
	slug := c.Params("slug")
	search := c.Query("search")
	filter := c.Query("filter")

	riders, err := h.service.ListRiders(c.Context(), slug, search, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": riders})
}

func (h *Handler) HandleVerifyStudent(c *fiber.Ctx) error {
	slug := c.Params("slug")
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var req struct {
		Verified bool `json:"verified"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.service.VerifyStudent(c.Context(), slug, userID, req.Verified); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "อัปเดตการยืนยันตัวตนนศ. เรียบร้อยแล้ว"})
}

func (h *Handler) HandleUpdateStatus(c *fiber.Ctx) error {
	slug := c.Params("slug")
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.service.UpdateUserStatus(c.Context(), slug, userID, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "อัปเดตสถานะบัญชีผู้ใช้เรียบร้อยแล้ว"})
}

func (h *Handler) HandleListDisputes(c *fiber.Ctx) error {
	slug := c.Params("slug")
	status := c.Query("status")

	disputes, err := h.service.ListDisputes(c.Context(), slug, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": disputes})
}

func (h *Handler) HandleProcessRefund(c *fiber.Ctx) error {
	slug := c.Params("slug")

	var req RefundPenaltyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.service.ProcessRefund(c.Context(), slug, req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "ดำเนินการคืนเงินค่าปรับเข้าวอลเล็ตสำเร็จแล้ว"})
}

