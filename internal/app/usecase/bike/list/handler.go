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

// @Summary รายการรถจักรยาน
// @Description ดึงข้อมูลรถจักรยานทั้งหมด (ถ้าเป็น User ทั่วไปจะเห็นเฉพาะรถที่ว่าง, ถ้าเป็น Staff/Admin จะเห็นทั้งหมด)
// @Tags Bikes
// @Produce json
// @Security BearerAuth
// @Success 200 {array} entity.BikeData "รายการรถจักรยาน"
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

