package rank

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler interface {
	Handle(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// @Summary ข้อมูลระดับสมาชิก (Rank)
// @Description ดึงข้อมูลระดับสมาชิกปัจจุบันและยอดการใช้งานสะสมย้อนหลัง 4 เดือนเพื่อคำนวณลำดับ
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response "ดึงข้อมูลสำเร็จ"
// @Router /v1/users/rank [get]
func (h *handler) Handle(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	uID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token payload"})
	}

	resp, err := h.service.GetUserRank(c.Context(), uID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}
