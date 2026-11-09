package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/org/manage"
	"github.com/gofiber/fiber/v2"
)

func SetupOrgRoutes(app *fiber.App, orgHandler manage.Handler) {
	org := app.Group("/api/v1/org/:orgSlug", middleware.AuthMiddleware(), middleware.OrgAdminOrSuper())

	org.Get("/overview", orgHandler.HandleOverview)
	org.Post("/campuses", orgHandler.HandleAddCampus)
}
