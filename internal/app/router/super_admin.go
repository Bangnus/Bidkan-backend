package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/bike_assign"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_list"
	"github.com/gofiber/fiber/v2"
)

func SetupSuperAdminRoutes(
	app *fiber.App,
	createTenantHandler tenant_create.Handler,
	listTenantHandler tenant_list.Handler,
	assignBikeHandler bike_assign.Handler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")
	super := v1.Group("/super")

	super.Post("/tenants", createTenantHandler.Handle)
	super.Get("/tenants", listTenantHandler.Handle)
	super.Post("/bikes/assign", assignBikeHandler.Handle)
}

