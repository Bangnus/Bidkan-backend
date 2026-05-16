package update_token

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

// @Summary อัปเดต FCM Token
// @Description อัปเดต Device Token (Firebase) สำหรับรับการแจ้งเตือนแบบ Push Notification บนมือถือ
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "รหัส Token ของอุปกรณ์"
// @Success 200 {object} map[string]string "อัปเดตสำเร็จ"
// @Router /v1/notifications/fcm-token [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if err := h.svc.Execute(c.Context(), userID, req.FCMToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "token updated successfully"})
}
