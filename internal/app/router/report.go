package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/report/summary"
	"github.com/gofiber/fiber/v2"
)

func SetupReportRoutes(
	app *fiber.App,
	summaryHandler summary.Handler,
) {
	// ปกป้องด้วย AuthMiddleware และจะไปเช็ค role == "admin" ข้างใน Handler
	v1 := app.Group("/api/v1/reports", middleware.AuthMiddleware())
	
	v1.Get("/summary", summaryHandler.Handle)
}
