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

// @Summary สร้างพื้นที่บริการใหม่ (Admin Only)
// @Description สร้างจุดจอดรถ (P) หรือ ขอบเขตพื้นที่ให้บริการใหม่ โดยระบุพิกัดและรัศมี
// @Tags Zones
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "ข้อมูลพื้นที่ที่ต้องการสร้าง"
// @Success 201 {object} map[string]interface{} "สร้างสำเร็จ"
// @Router /v1/zones [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.service.Create(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Zone created successfully"})
}
