package login

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type Handler interface {
	Handle(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// @Summary เข้าสู่ระบบ
// @Description เข้าสู่ระบบด้วยชื่อผู้ใช้และรหัสผ่านเพื่อรับ JWT Token สำหรับการเข้าถึง API อื่นๆ
// @Tags Users
// @Accept json
// @Produce json
// @Param request body Request true "ข้อมูลการเข้าสู่ระบบ"
// @Success 200 {object} Response "เข้าสู่ระบบสำเร็จ"
// @Failure 401 {object} map[string]interface{} "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง"
// @Router /v1/users/login [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	res, err := h.service.Login(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}
