package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/report/summary"
	"github.com/gofiber/fiber/v2"
)

func SetupReportRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	summaryHandler summary.Handler,
) {
	v1 := app.Group(
		"/api/v1/reports",
		authMW,
		middleware.RequireScope(security.ScopeBackoffice),
		middleware.BackofficeRoleGuard("super_admin", "org_admin", "tenant_admin"),
	)

	v1.Get("/summary", summaryHandler.Handle)
}
