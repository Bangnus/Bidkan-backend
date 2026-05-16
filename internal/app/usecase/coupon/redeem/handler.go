package redeem

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

// @Summary ใช้โค้ดคูปอง
// @Description กรอกโค้ดเพื่อรับเงินฟรีเข้า Wallet หรือเก็บโค้ดส่วนลดไว้ใช้หักค่าขี่ในอนาคต
// @Tags Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "โค้ดคูปองที่ต้องการใช้"
// @Success 200 {object} map[string]string "ใช้คูปองสำเร็จ"
// @Failure 400 {object} map[string]string "โค้ดไม่ถูกต้อง หรือหมดอายุ หรือถูกใช้ไปแล้ว"
// @Failure 500 {object} map[string]string "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /v1/coupons/redeem [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	userID, ok := c.Locals("userId").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	req.UserID = userID

	if err := h.svc.Execute(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "coupon redeemed successfully"})
}
