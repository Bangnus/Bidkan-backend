package end

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
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

// @Summary จบการเช่ารถจักรยาน
// @Description จบการขี่จักรยาน คำนวณค่าบริการ หักเงินใน Wallet และคืนสถานะรถให้ว่าง
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "ข้อมูลการจบการเช่า"
// @Success 200 {object} entity.Ride "จบการเช่าและชำระเงินสำเร็จ"
// @Router /v1/rides/end [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	userID, ok := middleware.GetAuthUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	req.UserID = userID

	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	ride, err := h.service.End(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(ride)
}
