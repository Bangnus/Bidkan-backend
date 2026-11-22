package app

import (
	"context"
	"database/sql"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	infraMqtt "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/Bangnus/Bidkan-backend/internal/app/router"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/auth"
	campusAudit "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/audit_manage"
	campusBikeList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_list"
	campusBikeManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_manage"
	campusCouponManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/coupon_manage"
	campusDashboard "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/dashboard"
	campusFieldOps "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/field_ops"
	campusIncident "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/incident_manage"
	campusInventory "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/inventory_manage"
	campusOrderManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/order_manage"
	campusRebalance "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/rebalancing"
	campusRemoteOps "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/remote_ops"
	campusReportManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/report_manage"
	campusRiderManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/rider_manage"
	campusRevenueManage "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/revenue_manage"
	campusRideList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/ride_list"
	campusStaffComp "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/staff_compensation"
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
	OrgRepo                 domainRepo.OrganizationRepository
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
	CampusStaffCompHandler  campusStaffComp.Handler
	CampusRemoteOpsHandler   campusRemoteOps.Handler
	CampusRiderHandler       campusRiderManage.Handler
	CampusRebalanceHandler   campusRebalance.Handler
	CampusAuditHandler       campusAudit.Handler
	CampusInventoryHandler   *campusInventory.Handler
	CampusIncidentHandler    *campusIncident.Handler
}

func NewTenantModule(db *sql.DB, secCfg *security.Config, sessions security.SessionStore, stripeClient infraStripe.Client, mqttPub infraMqtt.Publisher) *TenantModule {
	if stripeClient == nil {
		stripeClient = infraStripe.NewClient(nil)
	}
	userRepo := repository.NewUserPostgresRepository(db)
	orgRepo := repository.NewOrganizationPostgresRepository(db)
	tenantRepo := repository.NewTenantPostgresRepository(db)
	orderRepo := repository.NewHardwareOrderPostgresRepository(db)
	superRepo := repository.NewSuperAdminPostgresRepository(db)
	campusRepo := repository.NewCampusOpsPostgresRepository(db)
	staffCompRepo := repository.NewStaffCompensationPostgres(db)
	bikeRepo := repository.NewBikePostgresRepository(db)
	rideRepo := repository.NewRidePostgresRepository(db)
	riderRepo := repository.NewCampusRiderPostgres(db)
	rebalRepo := repository.NewRebalancePostgres(db)
	auditRepo := repository.NewAuditPostgres(db)
	partRepo := repository.NewSparePartPostgres(db)
	incRepo := repository.NewIncidentPostgres(db)

	return &TenantModule{
		TenantRepo:              tenantRepo,
		OrgRepo:                 orgRepo,
		AuthHandler:             auth.NewHandler(userRepo, tenantRepo, orgRepo, secCfg, sessions),
		PartnerApplyHandler:     partner_apply.NewHandler(orgRepo, tenantRepo, campusRepo, orderRepo),
		OrgManageHandler:        orgManage.NewHandler(orgRepo, tenantRepo, campusRepo),
		CreateTenantHandler:     superTenantCreate.NewHandler(superTenantCreate.NewService(tenantRepo, orgRepo, campusRepo)),
		ListTenantHandler:       superTenantList.NewHandler(superTenantList.NewService(tenantRepo)),
		ManageTenantHandler:     superTenantManage.NewHandler(tenantRepo, superRepo, orgRepo, campusRepo, orderRepo),
		AssignBikeHandler:       superBikeAssign.NewHandler(superBikeAssign.NewService(tenantRepo)),
		ManageBikeHandler:       superBikeManage.NewHandler(superRepo, tenantRepo),
		PayoutManageHandler:     superPayoutManage.NewHandler(superRepo, tenantRepo, campusRepo, stripeClient),
		CouponManageHandler:     superCouponManage.NewHandler(superRepo),
		CampusDashboardHandler:  campusDashboard.NewHandler(campusDashboard.NewService(tenantRepo, campusRepo)),
		CampusBikeListHandler:   campusBikeList.NewHandler(campusBikeList.NewService(tenantRepo)),
		CampusBikeManageHandler: campusBikeManage.NewHandler(campusRepo),
		CampusZoneHandler:       campusZoneManage.NewHandler(tenantRepo, campusRepo),
		CampusRideListHandler:   campusRideList.NewHandler(tenantRepo, campusRepo),
		CampusRevenueHandler:    campusRevenueManage.NewHandler(tenantRepo, campusRepo, stripeClient),
		CampusStaffHandler:      campusStaffManage.NewHandler(tenantRepo, campusRepo),
		CampusCouponHandler:     campusCouponManage.NewHandler(tenantRepo, campusRepo),
		CampusOrderHandler:      campusOrderManage.NewHandler(tenantRepo, orderRepo),
		CampusFieldOpsHandler:   campusFieldOps.NewHandler(campusFieldOps.NewService(tenantRepo, staffCompRepo)),
		CampusReportHandler:     campusReportManage.NewHandler(tenantRepo, campusRepo, staffCompRepo),
		CampusStaffCompHandler:  campusStaffComp.NewHandler(campusStaffComp.NewService(tenantRepo, staffCompRepo)),
		CampusRemoteOpsHandler:  campusRemoteOps.NewHandler(campusRemoteOps.NewService(tenantRepo, bikeRepo, rideRepo, mqttPub)),
		CampusRiderHandler:      campusRiderManage.NewHandler(campusRiderManage.NewService(tenantRepo, riderRepo)),
		CampusRebalanceHandler:  campusRebalance.NewHandler(campusRebalance.NewService(tenantRepo, campusRepo, rebalRepo)),
		CampusAuditHandler:      campusAudit.NewHandler(campusAudit.NewService(tenantRepo, campusRepo, staffCompRepo, auditRepo)),
		CampusInventoryHandler:  campusInventory.NewHandler(campusInventory.NewService(tenantRepo, partRepo, auditRepo)),
		CampusIncidentHandler:   campusIncident.NewHandler(campusIncident.NewService(tenantRepo, incRepo, auditRepo)),
	}
}

func (m *TenantModule) SeedAndRegisterRoutes(server *fiber.App, db *sql.DB, authMW fiber.Handler, sessions security.SessionStore) {
	auth.SeedDemoData(context.Background(), db)
	router.SetupAuthRoutes(server, authMW, sessions, m.AuthHandler, m.PartnerApplyHandler)
	router.SetupOrgRoutes(server, authMW, m.OrgRepo, m.OrgManageHandler)
	router.SetupSuperAdminRoutes(
		server,
		authMW,
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
		authMW,
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
		m.CampusStaffCompHandler,
		m.CampusRemoteOpsHandler,
		m.CampusRiderHandler,
		m.CampusRebalanceHandler,
		m.CampusAuditHandler,
		m.CampusInventoryHandler,
		m.CampusIncidentHandler,
	)
}
