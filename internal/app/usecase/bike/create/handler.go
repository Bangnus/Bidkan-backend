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

// @Summary เพิ่มรถจักรยานใหม่ (Admin Only)
// @Description เพิ่มรถจักรยานใหม่เข้าสู่ระบบ พร้อมระบุไอดี และสถานะเริ่มต้น
// @Tags Bikes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "ข้อมูลรถจักรยานใหม่"
// @Success 201 {object} map[string]interface{} "เพิ่มรถสำเร็จ"
// @Router /v1/bikes [post]
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

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Bike created successfully"})
}
