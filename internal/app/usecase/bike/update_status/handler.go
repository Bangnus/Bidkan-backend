package update_status

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

// @Summary Update Bike Status
// @Description Update bike status manually (e.g., to maintenance). Staff/Admin Only.
// @Tags Bikes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bike ID"
// @Param request body Request true "New Status"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /v1/bikes/{id}/status [patch]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// รับ ID จาก URL Param ถ้าใน Body ไม่ระบุ
	if req.BikeID == "" {
		req.BikeID = c.Params("id")
	}

	if req.BikeID == "" || req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bike_id and status are required"})
	}

	// เช็คสิทธิ์ (พนักงาน หรือ แอดมิน เท่านั้น)
	role := c.Locals("role").(string)
	if role != "staff" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "permission denied: staff or admin only"})
	}

	if err := h.svc.Execute(c.Context(), req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "bike status updated successfully",
		"bike_id": req.BikeID,
		"status":  req.Status,
	})
}
