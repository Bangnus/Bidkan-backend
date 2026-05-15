package list

import (
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	HandleListAvailable(c *fiber.Ctx) error
	HandleListAll(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// @Summary List Available Bikes
// @Description Get a list of bikes that are available for rent.
// @Tags Bikes
// @Produce json
// @Success 200 {array} entity.BikeData
// @Router /v1/bikes/available [get]
func (h *handler) HandleListAvailable(c *fiber.Ctx) error {
	bikes, err := h.service.GetAvailableBikes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(bikes)
}

// @Summary List All Bikes (Admin)
// @Description Get a list of all bikes in the system.
// @Tags Bikes
// @Produce json
// @Success 200 {array} entity.BikeData
// @Router /v1/bikes [get]
func (h *handler) HandleListAll(c *fiber.Ctx) error {
	bikes, err := h.service.GetAllBikes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(bikes)
}
