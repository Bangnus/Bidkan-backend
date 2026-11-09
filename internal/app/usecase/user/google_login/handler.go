package google_login

import (
	"github.com/gofiber/fiber/v2"
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

// @Summary เข้าสู่ระบบด้วย Google (ทั้งอีเมลมหาวิทยาลัยและบุคคลทั่วไป)
// @Description รับ credential (Google ID Token) ตรวจสอบและบันทึกลงฐานข้อมูล พร้อมแยกมหาวิทยาลัยอัตโนมัติ
// @Tags Users
// @Accept json
// @Produce json
// @Param request body Request true "Google OAuth Credential"
// @Success 200 {object} Response "เข้าสู่ระบบสำเร็จ พร้อม JWT Token"
// @Failure 400 {object} map[string]string "ข้อมูลไม่ถูกต้อง"
// @Failure 500 {object} map[string]string "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /v1/users/google-login [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	res, err := h.service.Login(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

