package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/topup"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/transfer"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/verify_receiver"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/webhook"

	"github.com/gofiber/fiber/v2"
)

func SetupWalletRoutes(
	app *fiber.App,
	topupHandler topup.Handler,
	webhookHandler webhook.Handler,
	transferHandler transfer.Handler,
	verifyHandler verify_receiver.Handler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Protected Routes
	walletGroup := v1.Group("/wallet", middleware.AuthMiddleware())
	
	walletGroup.Post("/topup", topupHandler.Handle)
	walletGroup.Post("/transfer", transferHandler.Handle)
	walletGroup.Get("/verify-receiver", verifyHandler.Handle)

	// Public Webhook (แยกออกมาเพราะไม่ต้องใช้ Auth)
	v1.Post("/wallet/webhook/paysolutions", webhookHandler.Handle)
}
