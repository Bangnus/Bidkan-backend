package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/update_status"
	"github.com/gofiber/fiber/v2"
)

func SetupBikeRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	optAuthMW fiber.Handler,
	createHandler create.Handler,
	listHandler list.Handler,
	updateStatusHandler update_status.Handler,
) {
	v1 := app.Group("/api/v1/bikes")

	// Public / Guest / User สามารถดึงดูรายการรถบนแผนที่ได้
	v1.Get("/", optAuthMW, listHandler.HandleList)

	// Protected สำหรับ Backoffice จัดการรถ
	protected := v1.Group("", authMW, middleware.RequireScope(security.ScopeBackoffice))
	protected.Post("/", middleware.SuperAdminOnly(), createHandler.Handle)
	protected.Patch("/:id/status", middleware.TenantStaffOrAdmin(), updateStatusHandler.Handle)
}
