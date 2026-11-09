package app

import (
	"context"
	"database/sql"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/router"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/auth"
	campusBikeList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_list"
	campusBikeManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_manage"
	campusCouponManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/coupon_manage"
	campusDashboard "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/dashboard"
	campusFieldOps "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/field_ops"
	campusOrderManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/order_manage"
	campusReportManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/report_manage"
	campusRevenueManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/revenue_manage"
	campusRideList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/ride_list"
	campusStaffManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/staff_manage"
	campusZoneManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/zone_manage"
	orgManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/org/manage"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/public/partner_apply"
	superBikeAssign "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/bike_assign"
	superBikeManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/bike_manage"
	superCouponManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/coupon_manage"
	superPayoutManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/payout_manage"
	superTenantCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_create"
	superTenantList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_list"
	superTenantManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_manage"
	"github.com/gofiber/fiber/v2"
)

type TenantModule struct {
	TenantRepo              domainRepo.TenantRepository
	AuthHandler             auth.Handler
	PartnerApplyHandler     partner_apply.Handler
	OrgManageHandler        orgManage.Handler
	CreateTenantHandler     superTenantCreate.Handler
	ListTenantHandler       superTenantList.Handler
	ManageTenantHandler     superTenantManage.Handler
	AssignBikeHandler       superBikeAssign.Handler
	ManageBikeHandler       superBikeManage.Handler
	PayoutManageHandler     superPayoutManage.Handler
	CouponManageHandler     superCouponManage.Handler
	CampusDashboardHandler  campusDashboard.Handler
	CampusBikeListHandler   campusBikeList.Handler
	CampusBikeManageHandler campusBikeManage.Handler
	CampusZoneHandler       campusZoneManage.Handler
	CampusRideListHandler   campusRideList.Handler
	CampusRevenueHandler    campusRevenueManage.Handler
	CampusStaffHandler      campusStaffManage.Handler
	CampusCouponHandler     campusCouponManage.Handler
	CampusOrderHandler      campusOrderManage.Handler
	CampusFieldOpsHandler   campusFieldOps.Handler
	CampusReportHandler     campusReportManage.Handler
}

func NewTenantModule(db *sql.DB) *TenantModule {
	userRepo := repository.NewUserPostgresRepository(db)
	orgRepo := repository.NewOrganizationPostgresRepository(db)
	tenantRepo := repository.NewTenantPostgresRepository(db)
	orderRepo := repository.NewHardwareOrderPostgresRepository(db)
	superRepo := repository.NewSuperAdminPostgresRepository(db)
	campusRepo := repository.NewCampusOpsPostgresRepository(db)

	return &TenantModule{
		TenantRepo:              tenantRepo,
		AuthHandler:             auth.NewHandler(userRepo, tenantRepo, orgRepo),
		PartnerApplyHandler:     partner_apply.NewHandler(orgRepo, tenantRepo, campusRepo, orderRepo),
		OrgManageHandler:        orgManage.NewHandler(orgRepo, tenantRepo, campusRepo),
		CreateTenantHandler:     superTenantCreate.NewHandler(superTenantCreate.NewService(tenantRepo, orgRepo, campusRepo)),
		ListTenantHandler:       superTenantList.NewHandler(superTenantList.NewService(tenantRepo)),
		ManageTenantHandler:     superTenantManage.NewHandler(tenantRepo, superRepo, orgRepo, campusRepo, orderRepo),
		AssignBikeHandler:       superBikeAssign.NewHandler(superBikeAssign.NewService(tenantRepo)),
		ManageBikeHandler:       superBikeManage.NewHandler(superRepo, tenantRepo),
		PayoutManageHandler:     superPayoutManage.NewHandler(superRepo, tenantRepo, campusRepo),
		CouponManageHandler:     superCouponManage.NewHandler(superRepo),
		CampusDashboardHandler:  campusDashboard.NewHandler(campusDashboard.NewService(tenantRepo, campusRepo)),
		CampusBikeListHandler:   campusBikeList.NewHandler(campusBikeList.NewService(tenantRepo)),
		CampusBikeManageHandler: campusBikeManage.NewHandler(campusRepo),
		CampusZoneHandler:       campusZoneManage.NewHandler(tenantRepo, campusRepo),
		CampusRideListHandler:   campusRideList.NewHandler(tenantRepo, campusRepo),
		CampusRevenueHandler:    campusRevenueManage.NewHandler(tenantRepo, campusRepo),
		CampusStaffHandler:      campusStaffManage.NewHandler(tenantRepo, campusRepo),
		CampusCouponHandler:     campusCouponManage.NewHandler(tenantRepo, campusRepo),
		CampusOrderHandler:      campusOrderManage.NewHandler(tenantRepo, orderRepo),
		CampusFieldOpsHandler:   campusFieldOps.NewHandler(campusFieldOps.NewService(tenantRepo)),
		CampusReportHandler:     campusReportManage.NewHandler(tenantRepo, campusRepo),
	}
}

func (m *TenantModule) SeedAndRegisterRoutes(server *fiber.App, db *sql.DB) {
	auth.SeedDemoData(context.Background(), db)
	router.SetupAuthRoutes(server, m.AuthHandler, m.PartnerApplyHandler)
	router.SetupOrgRoutes(server, m.OrgManageHandler)
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
		m.CampusCouponHandler,
		m.CampusOrderHandler,
		m.CampusFieldOpsHandler,
		m.CampusReportHandler,
	)
}
