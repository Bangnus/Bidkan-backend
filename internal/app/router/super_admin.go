package router

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/bike_assign"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/bike_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/coupon_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/payout_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_create"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_manage"
	"github.com/gofiber/fiber/v2"
)

func SetupSuperAdminRoutes(
	app *fiber.App,
	createTenantHandler tenant_create.Handler,
	listTenantHandler tenant_list.Handler,
	manageTenantHandler tenant_manage.Handler,
	assignBikeHandler bike_assign.Handler,
	manageBikeHandler bike_manage.Handler,
	payoutHandler payout_manage.Handler,
	couponHandler coupon_manage.Handler,
) {
	super := app.Group("/api/v1/super", middleware.AuthMiddleware(), middleware.SuperAdminOnly())

	// Tenants
	super.Get("/tenants", listTenantHandler.Handle)
	super.Post("/tenants", createTenantHandler.Handle)
	super.Patch("/tenants/:id/status", manageTenantHandler.HandleUpdateStatus)
	super.Patch("/tenants/:id/rates", manageTenantHandler.HandleUpdateRates)

	// Bikes Fleet Provisioning
	super.Get("/bikes", manageBikeHandler.HandleListAll)
	super.Post("/bikes", manageBikeHandler.HandleCreate)
	super.Post("/bikes/assign", assignBikeHandler.Handle)

	// Stripe Payouts
	super.Get("/payouts", payoutHandler.HandleListAll)
	super.Post("/payouts/trigger", payoutHandler.HandleTriggerPayout)

	// Global Coupons
	super.Get("/coupons", couponHandler.HandleList)
	super.Post("/coupons", couponHandler.HandleCreate)
}
