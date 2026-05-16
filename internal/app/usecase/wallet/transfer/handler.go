package transfer

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

// @Summary โอนเงินให้เพื่อน (P2P)
// @Description โอนเงินจาก Wallet ของเราไปยังผู้ใช้อื่นโดยระบุเบอร์โทรศัพท์ผู้รับ
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "ข้อมูลการโอนเงิน"
// @Success 200 {object} map[string]interface{} "โอนเงินสำเร็จ"
// @Failure 400 {object} map[string]string "ข้อมูลไม่ถูกต้อง หรือยอดเงินไม่พอ"
// @Failure 500 {object} map[string]string "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /v1/wallet/transfer [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	senderID, ok := c.Locals("userId").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	req.SenderID = senderID

	if err := h.svc.Execute(c.Context(), req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "transfer completed successfully",
		"amount":  req.Amount,
		"to":      req.ReceiverPhone,
	})
}
