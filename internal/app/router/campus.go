package router

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	campusAudit "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/audit_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/coupon_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/dashboard"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/field_ops"
	campusInventory "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/inventory_manage"
	campusIncident "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/incident_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/order_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/rebalancing"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/remote_ops"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/report_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/revenue_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/ride_list"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/rider_manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/staff_compensation"
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
	staffCompHandler staff_compensation.Handler,
	remoteOpsHandler remote_ops.Handler,
	riderHandler rider_manage.Handler,
	rebalanceHandler rebalancing.Handler,
	auditHandler campusAudit.Handler,
	inventoryHandler *campusInventory.Handler,
	incidentHandler *campusIncident.Handler,
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
	campus.Get("/staff/compensation-config", adminOnly, staffCompHandler.HandleGetConfig)
	campus.Put("/staff/compensation-config", adminOnly, staffCompHandler.HandleUpdateConfig)
	campus.Get("/staff/earnings", adminOnly, staffCompHandler.HandleListEarnings)
	campus.Get("/coupons", adminOnly, couponHandler.HandleList)
	campus.Post("/coupons", adminOnly, couponHandler.HandleCreate)
	campus.Get("/orders", adminOnly, orderHandler.HandleList)
	campus.Post("/orders", adminOnly, orderHandler.HandleCreate)

	// Rider & Student Management & Dispute Desk (Admin Only)
	campus.Get("/riders", adminOnly, riderHandler.HandleListRiders)
	campus.Patch("/riders/:id/verify-student", adminOnly, riderHandler.HandleVerifyStudent)
	campus.Patch("/riders/:id/status", adminOnly, riderHandler.HandleUpdateStatus)
	campus.Get("/disputes", adminOnly, riderHandler.HandleListDisputes)
	campus.Post("/disputes/refund", adminOnly, riderHandler.HandleProcessRefund)

	// Staff & Admin Routes (Fleet Status, Field Ops & Personal Earnings)
	campus.Get("/bikes", staffOrAdmin, bikeListHandler.Handle)
	campus.Patch("/bikes/:id/status", staffOrAdmin, bikeManageHandler.HandleUpdateStatus)
	campus.Get("/field/queue", staffOrAdmin, fieldOpsHandler.HandleGetQueue)
	campus.Post("/field/swap-battery", staffOrAdmin, fieldOpsHandler.HandleSwapBattery)
	campus.Get("/field/reports", staffOrAdmin, reportHandler.HandleList)
	campus.Patch("/field/reports/:id/resolve", staffOrAdmin, reportHandler.HandleResolve)
	campus.Get("/field/my-earnings", staffOrAdmin, staffCompHandler.HandleGetMyEarnings)

	// Live Fleet Operations & Remote IoT Controls
	campus.Get("/live/fleet", staffOrAdmin, remoteOpsHandler.HandleGetLiveFleet)
	campus.Post("/bikes/:id/remote/beep", staffOrAdmin, remoteOpsHandler.HandleRemoteBeep)
	campus.Post("/bikes/:id/remote/unlock", adminOnly, remoteOpsHandler.HandleRemoteUnlock)
	campus.Post("/bikes/:id/remote/lock", adminOnly, remoteOpsHandler.HandleRemoteLock)
	campus.Post("/bikes/:id/remote/force-end", adminOnly, remoteOpsHandler.HandleForceEnd)
	// Fleet Rebalancing & Demand Heatmap
	campus.Get("/rebalancing/overview", staffOrAdmin, rebalanceHandler.HandleGetOverview)

	// Accounting Export & Audit Activity Logs (Admin Only)
	campus.Get("/audit/logs", adminOnly, auditHandler.HandleListAuditLogs)
	campus.Get("/export/rides", adminOnly, auditHandler.HandleExportRidesCSV)
	campus.Get("/export/revenue", adminOnly, auditHandler.HandleExportRevenueCSV)
	campus.Get("/export/staff-payouts", adminOnly, auditHandler.HandleExportStaffPayoutsCSV)

	// Spare Parts & Battery Inventory Ledger (Staff & Admin)
	campus.Get("/inventory", staffOrAdmin, inventoryHandler.GetInventory)
	campus.Post("/inventory/adjust", staffOrAdmin, inventoryHandler.AdjustStock)
	campus.Post("/inventory/parts", adminOnly, inventoryHandler.CreatePart)
	campus.Get("/inventory/logs", staffOrAdmin, inventoryHandler.ListLogs)

	// Incident & Insurance Claim Desk (Staff & Admin)
	campus.Get("/incidents", adminOnly, incidentHandler.ListIncidents)
	campus.Post("/incidents", staffOrAdmin, incidentHandler.CreateIncident)
	campus.Patch("/incidents/:id/status", adminOnly, incidentHandler.UpdateStatus)
	campus.Get("/export/incidents", adminOnly, incidentHandler.ExportCSV)
}
