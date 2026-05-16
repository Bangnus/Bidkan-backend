package create

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

// @Summary สร้างคูปองใหม่ (Admin Only)
// @Description แอดมินสร้างคูปองใหม่เพื่อแจกจ่ายให้ผู้ใช้ โดยกำหนดรหัส (Code) ประเภท (Type) และมูลค่าได้
// @Tags Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "รายละเอียดคูปองที่ต้องการสร้าง"
// @Success 200 {object} map[string]string "สร้างคูปองสำเร็จ"
// @Failure 403 {object} map[string]string "ไม่มีสิทธิ์ (เฉพาะแอดมินเท่านั้น)"
// @Failure 500 {object} map[string]string "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /v1/coupons [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "permission denied: admin only"})
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	if err := h.svc.Execute(c.Context(), req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "coupon created successfully"})
}
