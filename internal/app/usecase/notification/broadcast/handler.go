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

// @Summary Broadcast Notification (Admin Only)
// @Description แอดมินส่งประกาศ Ads หรือข่าวสารหาผู้ใช้ทุกคน
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "เนื้อหาประกาศ"
// @Success 200 {object} map[string]string
// @Failure 403 {object} map[string]string
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
