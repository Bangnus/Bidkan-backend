package broadcast

import (
	"github.com/gofiber/fiber/v2"
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

// @Summary ส่งประกาศแจ้งเตือนหาทุกคน (Admin Only)
// @Description แอดมินส่งข่าวสาร โปรโมชั่น หรือประกาศสำคัญไปยังผู้ใช้ทุกคนผ่าน MQTT และ Push Notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "เนื้อหาที่ต้องการประกาศ"
// @Success 200 {object} map[string]string "ส่งประกาศสำเร็จ"
// @Failure 403 {object} map[string]string "ไม่มีสิทธิ์ (เฉพาะแอดมินเท่านั้น)"
// @Router /v1/admin/notifications/broadcast [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin only"})
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if err := h.svc.Execute(c.Context(), req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "broadcast sent successfully"})
}
