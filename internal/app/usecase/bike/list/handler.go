package list

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	HandleList(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// @Summary List Bikes
// @Description Get a list of bikes based on user role. Normal users see available bikes only. Staff/Admin see all.
// @Tags Bikes
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.BikeData
// @Router /v1/bikes [get]
func (h *handler) HandleList(c *fiber.Ctx) error {
	role, ok := c.Locals("role").(string)
	if !ok {
		role = "user" // Default to user if role is missing
	}

	var bikes []entity.BikeData
	var err error

	if role == "staff" || role == "admin" {
		// Staff and Admin see everything
		bikes, err = h.service.GetAllBikes(c.Context())
	} else {
		// Normal users see only available bikes
		bikes, err = h.service.GetAvailableBikes(c.Context())
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(bikes)
}

