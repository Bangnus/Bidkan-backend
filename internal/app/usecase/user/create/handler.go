package create

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

// Handle Create User
// @Summary Create a new user
// @Description Register a new user with email, name, and password
// @Tags Users
// @Accept json
// @Produce json
// @Param request body Request true "User creation request"
// @Router /users [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request

	// 1. Parse JSON Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	// (ควรมีขั้นตอน Validate Struct ที่นี่)

	// 2. เรียกใช้ Service (ส่ง Context ของ Fiber ไปด้วยเผื่อกรณี Request ถูก Cancel)
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
