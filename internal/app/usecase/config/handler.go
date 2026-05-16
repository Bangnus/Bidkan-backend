package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type Handler interface {
	HandleSet(c *fiber.Ctx) error
	HandleGet(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// @Summary ตั้งค่าระบบ (Admin Only)
// @Description ตั้งค่าพารามิเตอร์ต่างๆ ของระบบ (เช่น ค่าปรับจอดนอกพื้นที่, อัตราค่าบริการ)
// @Tags Configs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "ข้อมูลการตั้งค่า"
// @Success 200 {object} map[string]interface{} "ตั้งค่าสำเร็จ"
// @Router /v1/configs [post]
func (h *handler) HandleSet(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.service.SetConfig(c.Context(), req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Config updated successfully"})
}

// @Summary ดึงค่าการตั้งค่าระบบ
// @Description ดึงค่าการตั้งค่าตาม Key ที่ระบุ
// @Tags Configs
// @Produce json
// @Param key path string true "ชื่อของการตั้งค่า (Key)"
// @Success 200 {object} map[string]interface{} "ดึงข้อมูลสำเร็จ"
// @Router /v1/configs/{key} [get]
func (h *handler) HandleGet(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "key is required"})
	}

	val, err := h.service.GetConfig(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Config not found"})
	}

	return c.JSON(fiber.Map{"key": key, "value": val})
}
