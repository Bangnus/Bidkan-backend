package verify_firebase

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

// @Summary Verify Firebase Token and Activate User
// @Description Verify the idToken received from Firebase on Frontend to activate the user account.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body Request true "Firebase verification request"
// @Success 200 {object} map[string]interface{} "User activated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid token or request"
// @Router /v1/users/verify/firebase [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.service.VerifyAndActivate(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User activated via Firebase successfully.",
	})
}
