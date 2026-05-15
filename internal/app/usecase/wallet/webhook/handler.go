package webhook

import (
	"log"

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

// @Summary PaySolutions Webhook Callback
// @Description Receives payment status updates from PaySolutions server.
// @Tags Wallet
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/wallet/webhook/paysolutions [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var payload PaySolutionsWebhookPayload

	// PaySolutions อาจจะส่งมาเป็น Form-data หรือ JSON 
	// หากเป็น Form-data อาจจะต้องใช้ c.BodyParser หรือดึงค่าทีละตัว (สมมติว่าเป็น JSON ไปก่อน)
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("Webhook Error: invalid payload format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	// เรียก Service เพื่อประมวลผลยอดเงิน
	err := h.service.ProcessCallback(c.Context(), payload)
	if err != nil {
		log.Printf("Webhook Error: %v", err)
		// ถึงแม้จะ error ภายใน (เช่น หายอดไม่เจอ) ปกติมักจะคืน 200 กลับไปให้ Gateway เพื่อไม่ให้มันยิงซ้ำ
		// แต่เพื่อความปลอดภัย ให้คืน 200 หรือ 400 ตามความเหมาะสม
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "processed with error", "detail": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success"})
}
