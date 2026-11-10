package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/config"
	"github.com/gofiber/fiber/v2"
)

func SetupConfigRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	handler config.Handler,
) {
	v1 := app.Group("/api/v1/configs", authMW)

	// เฉพาะ Super Admin เท่านั้นที่แก้ค่าระบบกลางได้
	v1.Post("/", middleware.RequireScope(security.ScopeBackoffice), middleware.SuperAdminOnly(), handler.HandleSet)

	// ทุกคนที่มี Token ดูค่าได้
	v1.Get("/", handler.HandleList)
	v1.Get("/:key", handler.HandleGet)
}
