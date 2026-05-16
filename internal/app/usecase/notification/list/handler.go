package list

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler interface {
	Handle(c *fiber.Ctx) error
}

type handler struct {
	svc Service
}

func NewHandler(svc Service) Handler {
	return &handler{svc: svc}
}

// @Summary รายการแจ้งเตือนของฉัน
// @Description ดึงประวัติการแจ้งเตือนทั้งหมดของผู้ใช้ (รวมทั้งประกาศส่วนกลางและแจ้งเตือนส่วนตัว)
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {array} NotificationInfo "รายการแจ้งเตือน"
// @Router /v1/notifications [get]
func (h *handler) Handle(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	results, err := h.svc.Execute(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(results)
}
