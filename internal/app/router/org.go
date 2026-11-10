package router

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/org/manage"
	"github.com/gofiber/fiber/v2"
)

func SetupOrgRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	orgRepo domainRepo.OrganizationRepository,
	orgHandler manage.Handler,
) {
	org := app.Group(
		"/api/v1/org/:orgSlug",
		authMW,
		middleware.RequireScope(security.ScopeBackoffice),
		middleware.OrgAdminOnly(),
		middleware.OrgSlugGuard(orgRepo),
	)

	org.Get("/overview", orgHandler.HandleOverview)
	org.Post("/campuses", orgHandler.HandleAddCampus)
}
