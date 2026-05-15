package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/topup"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/wallet/webhook"

	"github.com/gofiber/fiber/v2"
)

func SetupWalletRoutes(
	app *fiber.App,
	topupHandler topup.Handler,
	webhookHandler webhook.Handler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Protected Routes
	walletGroup := v1.Group("/wallet")
	
	// API สำหรับขอ QR Code (ต้อง Login)
	walletGroup.Post("/topup", middleware.AuthMiddleware(), topupHandler.Handle)

	// Public Webhook (PaySolutions ยิงมา ดังนั้นไม่ต้องมี Auth ของแอป แต่เดี๋ยวต้องมีวิธีเช็ค Header/IP จากฝั่ง Payment Gateway)
	walletGroup.Post("/webhook/paysolutions", webhookHandler.Handle)
}
