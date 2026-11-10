package router

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/coupon_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/dashboard"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/field_ops"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/order_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/report_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/revenue_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/ride_list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/staff_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/zone_manage"
	"github.com/gofiber/fiber/v2"
)

func SetupCampusRoutes(
	app *fiber.App,
	authMW fiber.Handler,
	tenantRepo domainRepo.TenantRepository,
	dashboardHandler dashboard.Handler,
	bikeListHandler bike_list.Handler,
	bikeManageHandler bike_manage.Handler,
	zoneHandler zone_manage.Handler,
	rideListHandler ride_list.Handler,
	revenueHandler revenue_manage.Handler,
	staffHandler staff_manage.Handler,
	couponHandler coupon_manage.Handler,
	orderHandler order_manage.Handler,
	fieldOpsHandler field_ops.Handler,
	reportHandler report_manage.Handler,
) {
	campus := app.Group(
		"/api/v1/campus/:slug",
		authMW,
		middleware.RequireScope(security.ScopeBackoffice),
		middleware.CampusSlugGuard(tenantRepo),
	)
	adminOnly := middleware.TenantAdminOnly()
	staffOrAdmin := middleware.TenantStaffOrAdmin()

	// Admin Only Routes (Dashboard, Zones, Rides, Revenue, Staff, Coupons, Hardware Orders)
	campus.Get("/dashboard", adminOnly, dashboardHandler.Handle)
	campus.Get("/zones", adminOnly, zoneHandler.HandleList)
	campus.Post("/zones", adminOnly, zoneHandler.HandleCreate)
	campus.Delete("/zones/:id", adminOnly, zoneHandler.HandleDelete)
	campus.Get("/rides", adminOnly, rideListHandler.HandleList)
	campus.Get("/revenue", adminOnly, revenueHandler.HandleGetRevenue)
	campus.Patch("/rates", adminOnly, revenueHandler.HandleUpdateRates)
	campus.Get("/stripe/status", adminOnly, revenueHandler.HandleGetStripeStatus)
	campus.Post("/stripe/onboard", adminOnly, revenueHandler.HandleStripeOnboard)
	campus.Post("/stripe/login-link", adminOnly, revenueHandler.HandleStripeLoginLink)
	campus.Get("/staff", adminOnly, staffHandler.HandleList)
	campus.Post("/staff", adminOnly, staffHandler.HandleCreate)
	campus.Get("/coupons", adminOnly, couponHandler.HandleList)
	campus.Post("/coupons", adminOnly, couponHandler.HandleCreate)
	campus.Get("/orders", adminOnly, orderHandler.HandleList)
	campus.Post("/orders", adminOnly, orderHandler.HandleCreate)

	// Staff & Admin Routes (Fleet Status & Field Ops)
	campus.Get("/bikes", staffOrAdmin, bikeListHandler.Handle)
	campus.Patch("/bikes/:id/status", staffOrAdmin, bikeManageHandler.HandleUpdateStatus)
	campus.Get("/field/queue", staffOrAdmin, fieldOpsHandler.HandleGetQueue)
	campus.Post("/field/swap-battery", staffOrAdmin, fieldOpsHandler.HandleSwapBattery)
	campus.Get("/field/reports", staffOrAdmin, reportHandler.HandleList)
	campus.Patch("/field/reports/:id/resolve", staffOrAdmin, reportHandler.HandleResolve)
}
