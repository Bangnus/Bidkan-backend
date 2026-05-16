package topup

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

// @Summary ขอเติมเงินเข้า Wallet (PromptPay)
// @Description สร้างรายการธุรกรรมเติมเงินและรับ QR Code สำหรับชำระเงินผ่าน PromptPay
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Request true "ระบุจำนวนเงินที่ต้องการเติม"
// @Success 200 {object} Response "สร้างรายการเติมเงินสำเร็จ"
// @Router /v1/wallet/topup [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	uID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token payload"})
	}
	req.UserID = uID

	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	resp, err := h.service.RequestTopup(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}
