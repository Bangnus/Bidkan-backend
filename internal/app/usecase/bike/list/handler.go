package list

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
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
	role := middleware.GetAuthRole(c)
	scope := middleware.GetAuthScope(c)

	var bikes []entity.BikeData
	var err error

	isBackoffice := scope == "backoffice" && (role == "super_admin" || role == "org_admin" || role == "tenant_admin" || role == "tenant_staff")
	if isBackoffice {
		bikes, err = h.service.GetAllBikes(c.Context())
	} else {
		bikes, err = h.service.GetAvailableBikes(c.Context())
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(bikes)
}
