package router

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/dashboard"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/field_ops"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/report_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/revenue_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/ride_list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/staff_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/zone_manage"
	"github.com/gofiber/fiber/v2"
)

func SetupCampusRoutes(
	app *fiber.App,
	tenantRepo domainRepo.TenantRepository,
	dashboardHandler dashboard.Handler,
	bikeListHandler bike_list.Handler,
	bikeManageHandler bike_manage.Handler,
	zoneHandler zone_manage.Handler,
	rideListHandler ride_list.Handler,
	revenueHandler revenue_manage.Handler,
	staffHandler staff_manage.Handler,
	fieldOpsHandler field_ops.Handler,
	reportHandler report_manage.Handler,
) {
	campus := app.Group("/api/v1/campus/:slug", middleware.AuthMiddleware(), middleware.CampusSlugGuard(tenantRepo))

	// Admin Only Routes (Dashboard, Zones, Rides, Revenue, Staff)
	admin := campus.Group("", middleware.TenantAdminOnly())
	admin.Get("/dashboard", dashboardHandler.Handle)
	admin.Get("/zones", zoneHandler.HandleList)
	admin.Post("/zones", zoneHandler.HandleCreate)
	admin.Delete("/zones/:id", zoneHandler.HandleDelete)
	admin.Get("/rides", rideListHandler.HandleList)
	admin.Get("/revenue", revenueHandler.HandleGetRevenue)
	admin.Patch("/rates", revenueHandler.HandleUpdateRates)
	admin.Post("/stripe/onboard", revenueHandler.HandleStripeOnboard)
	admin.Get("/staff", staffHandler.HandleList)
	admin.Post("/staff", staffHandler.HandleCreate)

	// Staff & Admin Routes (Fleet Status & Field Ops)
	ops := campus.Group("", middleware.TenantStaffOrAdmin())
	ops.Get("/bikes", bikeListHandler.Handle)
	ops.Patch("/bikes/:id/status", bikeManageHandler.HandleUpdateStatus)
	ops.Get("/field/queue", fieldOpsHandler.HandleGetQueue)
	ops.Post("/field/swap-battery", fieldOpsHandler.HandleSwapBattery)
	ops.Get("/field/reports", reportHandler.HandleList)
	ops.Patch("/field/reports/:id/resolve", reportHandler.HandleResolve)
}
