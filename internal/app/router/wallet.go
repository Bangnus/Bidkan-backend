package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/stripe_topup"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/transfer"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/verify_receiver"

	"github.com/gofiber/fiber/v2"
)

func SetupWalletRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	stripeTopupHandler stripe_topup.Handler,
	transferHandler transfer.Handler,
	verifyHandler verify_receiver.Handler,
) {
	v1 := app.Group("/api/v1")

	// Protected Routes (Rider Scope - Stripe Checkout)
	riderOnly := middleware.RequireScope(security.ScopeRider)
	v1.Get("/wallet/topup/config", authMW, riderOnly, stripeTopupHandler.HandleConfig)
	v1.Post("/wallet/topup/stripe", authMW, riderOnly, stripeTopupHandler.HandleCheckout)
	v1.Get("/wallet/topup/stripe/verify", authMW, riderOnly, stripeTopupHandler.HandleVerify)
	v1.Post("/wallet/transfer", authMW, riderOnly, transferHandler.Handle)
	v1.Get("/wallet/verify-receiver", authMW, riderOnly, verifyHandler.Handle)
}
