package list

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

// @Summary รายการพื้นที่บริการ (Zones)
// @Description ดึงข้อมูลพื้นที่บริการทั้งหมด รวมถึงจุดจอด (Parking Zones) และพื้นที่ห้ามจอด
// @Tags Zones
// @Produce json
// @Success 200 {array} entity.Zone "รายการพื้นที่"
// @Router /v1/zones [get]
func (h *handler) Handle(c *fiber.Ctx) error {
	zones, err := h.service.GetAllZones(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(zones)
}
