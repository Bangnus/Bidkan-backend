package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/stripe_webhook"
	"github.com/gofiber/fiber/v2"
)

func SetupStripeWebhookRoutes(app *fiber.App, handler stripe_webhook.Handler) {
	app.Post("/api/v1/webhooks/stripe", handler.Handle)
}

