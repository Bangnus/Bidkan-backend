package verify

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

// @Summary Verify OTP and Activate User
// @Description Verify the OTP sent to user's phone to activate their account.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body Request true "Verification request"
// @Success 200 {object} map[string]interface{} "User activated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid OTP or request"
// @Router /v1/users/verify [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.service.VerifyOTP(c.Context(), req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User activated successfully. You can now login.",
	})
}
