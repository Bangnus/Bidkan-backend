package verify_receiver

import (
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	Handle(c *fiber.Ctx) error
}

type handler struct {
	svc Service
}

func NewHandler(svc Service) Handler {
	return &handler{svc: svc}
}

// @Summary Verify Receiver
// @Description Find username by phone number for wallet transfer.
// @Tags Wallet
// @Produce json
// @Security BearerAuth
// @Param phone query string true "Receiver Phone Number"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/wallet/verify-receiver [get]
func (h *handler) Handle(c *fiber.Ctx) error {
	phone := c.Query("phone")
	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "phone number is required"})
	}

	res, err := h.svc.Execute(c.Context(), phone)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}
