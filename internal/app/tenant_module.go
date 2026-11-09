package app

import (
	"context"
	"database/sql"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/router"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/auth"
	"github.com/gofiber/fiber/v2"
	campusBikeList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_list"
	campusBikeManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_manage"
	campusDashboard "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/dashboard"
	campusFieldOps "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/field_ops"
	campusReportManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/report_manage"
	campusRevenueManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/revenue_manage"
	campusRideList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/ride_list"
	campusStaffManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/staff_manage"
	campusZoneManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/zone_manage"
	superBikeAssign "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/bike_assign"
	superBikeManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/bike_manage"
	superCouponManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/coupon_manage"
	superPayoutManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/payout_manage"
	superTenantCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_create"
	superTenantList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_list"
	superTenantManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_manage"
)

type TenantModule struct {
	TenantRepo             domainRepo.TenantRepository
	AuthHandler            auth.Handler
	CreateTenantHandler    superTenantCreate.Handler
	ListTenantHandler      superTenantList.Handler
	ManageTenantHandler    superTenantManage.Handler
	AssignBikeHandler      superBikeAssign.Handler
	ManageBikeHandler      superBikeManage.Handler
	PayoutManageHandler    superPayoutManage.Handler
	CouponManageHandler    superCouponManage.Handler
	CampusDashboardHandler campusDashboard.Handler
	CampusBikeListHandler  campusBikeList.Handler
	CampusBikeManageHandler campusBikeManage.Handler
	CampusZoneHandler      campusZoneManage.Handler
	CampusRideListHandler  campusRideList.Handler
	CampusRevenueHandler   campusRevenueManage.Handler
	CampusStaffHandler     campusStaffManage.Handler
	CampusFieldOpsHandler  campusFieldOps.Handler
	CampusReportHandler    campusReportManage.Handler
}

func NewTenantModule(db *sql.DB) *TenantModule {
	userRepo := repository.NewUserPostgresRepository(db)
	tenantRepo := repository.NewTenantPostgresRepository(db)
	superRepo := repository.NewSuperAdminPostgresRepository(db)
	campusRepo := repository.NewCampusOpsPostgresRepository(db)

	return &TenantModule{
		TenantRepo:              tenantRepo,
		AuthHandler:             auth.NewHandler(userRepo, tenantRepo),
		CreateTenantHandler:     superTenantCreate.NewHandler(superTenantCreate.NewService(tenantRepo)),
		ListTenantHandler:       superTenantList.NewHandler(superTenantList.NewService(tenantRepo)),
		ManageTenantHandler:     superTenantManage.NewHandler(tenantRepo, superRepo),
		AssignBikeHandler:       superBikeAssign.NewHandler(superBikeAssign.NewService(tenantRepo)),
		ManageBikeHandler:       superBikeManage.NewHandler(superRepo, tenantRepo),
		PayoutManageHandler:     superPayoutManage.NewHandler(superRepo, tenantRepo),
		CouponManageHandler:     superCouponManage.NewHandler(superRepo),
		CampusDashboardHandler:  campusDashboard.NewHandler(campusDashboard.NewService(tenantRepo)),
		CampusBikeListHandler:   campusBikeList.NewHandler(campusBikeList.NewService(tenantRepo)),
		CampusBikeManageHandler: campusBikeManage.NewHandler(campusRepo),
		CampusZoneHandler:       campusZoneManage.NewHandler(tenantRepo, campusRepo),
		CampusRideListHandler:   campusRideList.NewHandler(tenantRepo, campusRepo),
		CampusRevenueHandler:    campusRevenueManage.NewHandler(tenantRepo, campusRepo),
		CampusStaffHandler:      campusStaffManage.NewHandler(tenantRepo, campusRepo),
		CampusFieldOpsHandler:   campusFieldOps.NewHandler(campusFieldOps.NewService(tenantRepo)),
		CampusReportHandler:     campusReportManage.NewHandler(tenantRepo, campusRepo),
	}
}

func (m *TenantModule) SeedAndRegisterRoutes(server *fiber.App, db *sql.DB) {
	auth.SeedDemoData(context.Background(), db)
	router.SetupAuthRoutes(server, m.AuthHandler)
	router.SetupSuperAdminRoutes(
		server,
		m.CreateTenantHandler,
		m.ListTenantHandler,
		m.ManageTenantHandler,
		m.AssignBikeHandler,
		m.ManageBikeHandler,
		m.PayoutManageHandler,
		m.CouponManageHandler,
	)
	router.SetupCampusRoutes(
		server,
		m.TenantRepo,
		m.CampusDashboardHandler,
		m.CampusBikeListHandler,
		m.CampusBikeManageHandler,
		m.CampusZoneHandler,
		m.CampusRideListHandler,
		m.CampusRevenueHandler,
		m.CampusStaffHandler,
		m.CampusFieldOpsHandler,
		m.CampusReportHandler,
	)
}
