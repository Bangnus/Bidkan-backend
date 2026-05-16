package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/broadcast"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/notification/update_token"
	"github.com/gofiber/fiber/v2"
)

func SetupNotificationRoutes(
	app *fiber.App,
	broadcastHandler broadcast.Handler,
	listHandler list.Handler,
	updateTokenHandler update_token.Handler,
) {
	v1 := app.Group("/api/v1", middleware.AuthMiddleware())
	
	// User Routes
	v1.Get("/notifications", listHandler.Handle)
	v1.Post("/notifications/fcm-token", updateTokenHandler.Handle)

	// Admin Routes
	admin := v1.Group("/admin")
	admin.Post("/notifications/broadcast", broadcastHandler.Handle)
}
