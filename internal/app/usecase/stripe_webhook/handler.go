package stripe_webhook

import (
	"encoding/json"
	"errors"
	"log"

	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
	stripe  infraStripe.Client
}

func NewHandler(service Service, stripeClient infraStripe.Client) Handler {
	if stripeClient == nil {
		stripeClient = infraStripe.NewClient(nil)
	}
	return Handler{service: service, stripe: stripeClient}
}

func (h *Handler) Handle(c *fiber.Ctx) error {
	secrets := h.stripe.Config().WebhookSecrets()
	err := infraStripe.VerifySignature(
		c.Body(),
		c.Get("Stripe-Signature"),
		infraStripe.DefaultWebhookTolerance,
		secrets...,
	)
	if err != nil {
		if errors.Is(err, infraStripe.ErrWebhookNotArmed) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Stripe webhook secret not configured",
				"code":  "STRIPE_WEBHOOK_NOT_CONFIGURED",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var evt infraStripe.Event
	if err := json.Unmarshal(c.Body(), &evt); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid stripe event JSON"})
	}

	if err := h.service.HandleEvent(c.Context(), evt); err != nil {
		log.Printf("stripe webhook %s (%s) failed: %v", evt.Type, evt.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "webhook processing failed"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"received": true})
}

