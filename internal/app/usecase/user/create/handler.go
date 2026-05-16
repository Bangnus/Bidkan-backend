package create

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

// Handle Create User
// @Summary สมัครสมาชิกใหม่ (และส่ง OTP)
// @Description ลงทะเบียนผู้ใช้ใหม่ด้วยเบอร์โทรศัพท์ ชื่อผู้ใช้ และรหัสผ่าน หลังจากสำเร็จจะส่งรหัส OTP ไปยังเบอร์ที่ระบุ
// @Tags Users
// @Accept json
// @Produce json
// @Param request body Request true "ข้อมูลการสมัครสมาชิก"
// @Success 201 {object} map[string]interface{} "สมัครสมาชิกสำเร็จ รอการยืนยัน OTP"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง หรือการตรวจสอบไม่ผ่าน"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /v1/users [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request

	// 1. Parse JSON Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	// 2. Validate Struct
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// 3. เรียกใช้ Service (ส่ง Context ของ Fiber ไปด้วยเผื่อกรณี Request ถูก Cancel)
	res, err := h.service.CreateUser(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. ส่ง Response เป็น JSON
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}
