package me

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

// @Summary ข้อมูลโปรไฟล์ส่วนตัว
// @Description ดึงข้อมูลโปรไฟล์ของผู้คนที่กำลังเข้าสู่ระบบอยู่โดยใช้ JWT Token
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response "ดึงข้อมูลสำเร็จ"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต (Token ไม่ถูกต้องหรือหมดอายุ)"
// @Router /v1/users/me [get]
func (h *handler) Handle(c *fiber.Ctx) error {
	// ดึง user_id จาก Locals ที่ Middleware ฝากไว้
	userID := c.Locals("user_id").(string)

	res, err := h.service.GetMyProfile(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}
