package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/stripe_topup"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/topup"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/transfer"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/verify_receiver"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/webhook"

	"github.com/gofiber/fiber/v2"
)

func SetupWalletRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	topupHandler topup.Handler,
	stripeTopupHandler stripe_topup.Handler,
	webhookHandler webhook.Handler,
	transferHandler transfer.Handler,
	verifyHandler verify_receiver.Handler,
) {
	v1 := app.Group("/api/v1")

	// Public Webhook (HMAC-SHA256 Signature verified inside Handler)
	v1.Post("/wallet/webhook/paysolutions", webhookHandler.Handle)

	// Protected Routes (Rider Scope)
	riderOnly := middleware.RequireScope(security.ScopeRider)
	v1.Get("/wallet/topup/config", authMW, riderOnly, stripeTopupHandler.HandleConfig)
	v1.Post("/wallet/topup", authMW, riderOnly, topupHandler.Handle)
	v1.Post("/wallet/topup/stripe", authMW, riderOnly, stripeTopupHandler.HandleCheckout)
	v1.Get("/wallet/topup/stripe/verify", authMW, riderOnly, stripeTopupHandler.HandleVerify)
	v1.Post("/wallet/transfer", authMW, riderOnly, transferHandler.Handle)
	v1.Get("/wallet/verify-receiver", authMW, riderOnly, verifyHandler.Handle)
}
