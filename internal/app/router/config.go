package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/config"
	"github.com/gofiber/fiber/v2"
)

func SetupConfigRoutes(
	app *fiber.App,
	handler config.Handler,
) {
	v1 := app.Group("/api/v1/configs", middleware.AuthMiddleware())

	// เฉพาะ Admin เท่านั้นที่แก้ค่าได้
	v1.Post("/", middleware.AdminOnly(), handler.HandleSet)
	
	// ทุกคนที่มี Token ดูค่าได้
	v1.Get("/", handler.HandleList)
	v1.Get("/:key", handler.HandleGet)
}
