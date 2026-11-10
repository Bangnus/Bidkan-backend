package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	Handle(c *fiber.Ctx) error
}

type handler struct {
	service       Service
	webhookSecret string
}

func NewHandler(service Service, webhookSecret string) Handler {
	return &handler{service: service, webhookSecret: strings.TrimSpace(webhookSecret)}
}

// @Summary PaySolutions Webhook (ระบบรับแจ้งชำระเงิน)
// @Description รับการแจ้งเตือนสถานะการชำระเงินจาก PaySolutions (ใช้สำหรับเติมเงินอัตโนมัติ)
// @Tags Wallet
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "รับข้อมูลสำเร็จ"
// @Router /v1/wallet/webhook/paysolutions [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	sigHeader := strings.TrimSpace(c.Get("X-PaySolutions-Signature"))
	if !verifyWebhookSignature(c.Body(), sigHeader, h.webhookSecret) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid webhook signature"})
	}

	var payload PaySolutionsWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("Webhook Error: invalid payload format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := h.service.ProcessCallback(c.Context(), payload); err != nil {
		log.Printf("Webhook Error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success"})
}

func verifyWebhookSignature(rawBody []byte, signatureHex, secret string) bool {
	if secret == "" || signatureHex == "" {
		return false
	}
	cleaned := strings.TrimPrefix(strings.ToLower(signatureHex), "sha256=")
	sigBytes, err := hex.DecodeString(cleaned)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expected := mac.Sum(nil)
	return hmac.Equal(sigBytes, expected)
}
