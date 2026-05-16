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

// @Summary ตรวจสอบผู้รับโอน
// @Description ค้นหาชื่อผู้ใช้งานจากเบอร์โทรศัพท์ เพื่อยืนยันตัวตนก่อนทำการโอนเงิน
// @Tags Wallet
// @Produce json
// @Security BearerAuth
// @Param phone query string true "เบอร์โทรศัพท์ผู้รับ"
// @Success 200 {object} Response "พบข้อมูลผู้รับ"
// @Failure 400 {object} map[string]string "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]string "ไม่พบผู้ใช้งานเบอร์นี้"
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
