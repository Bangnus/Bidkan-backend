package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/update_status"
	"github.com/gofiber/fiber/v2"
)

func SetupBikeRoutes(
	app *fiber.App,
	createHandler create.Handler,
	listHandler list.Handler,
	updateStatusHandler update_status.Handler,
) {
	v1 := app.Group("/api/v1/bikes")

	// Public / Guest / User สามารถดึงดูรายการรถบนแผนที่ได้
	v1.Get("/", middleware.OptionalAuthMiddleware(), listHandler.HandleList)

	// Protected สำหรับ Admin และ Staff จัดการรถ
	protected := v1.Group("", middleware.AuthMiddleware())
	protected.Post("/", createHandler.Handle)            // Admin Only
	protected.Patch("/:id/status", updateStatusHandler.Handle) // Staff/Admin Only
}
