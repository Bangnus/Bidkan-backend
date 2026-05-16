package start

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

// @Summary เริ่มการเช่ารถจักรยาน
// @Description เริ่มต้นการขี่จักรยานโดยระบุไอดีรถที่ต้องการ (สแกน QR Code)
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "ข้อมูลการเริ่มเช่า"
// @Success 201 {object} entity.Ride "เริ่มการเช่าสำเร็จ"
// @Router /v1/rides/start [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// ถ้าไม่ส่ง UserID มาใน Body ให้ดึงจาก Token (ความปลอดภัยสูงขึ้น)
	userIDStr, ok := c.Locals("user_id").(string)
	if ok {
		uID, _ := uuid.Parse(userIDStr)
		req.UserID = uID
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	ride, err := h.service.Start(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(ride)
}
