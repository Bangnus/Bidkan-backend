package list_my

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

// @Summary รายการคูปองของฉัน
// @Description ดึงรายการคูปองที่ผู้ใช้เก็บไว้และยังไม่ได้ใช้
// @Tags Coupons
// @Produce json
// @Security BearerAuth
// @Success 200 {array} CouponInfo
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/coupons/my [get]
func (h *handler) Handle(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	coupons, err := h.svc.Execute(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(coupons)
}
