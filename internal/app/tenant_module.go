package app

import (
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/repository"
	campusBikeList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/bike_list"
	campusDashboard "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/dashboard"
	campusFieldOps "github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/field_ops"
	superBikeAssign "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/bike_assign"
	superTenantCreate "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_create"
	superTenantList "github.com/Bangnus/Bidkan-backend/internal/app/usecase/super/tenant_list"
)

type TenantModule struct {
	CreateTenantHandler    superTenantCreate.Handler
	ListTenantHandler      superTenantList.Handler
	AssignBikeHandler      superBikeAssign.Handler
	CampusDashboardHandler campusDashboard.Handler
	CampusBikeListHandler  campusBikeList.Handler
	CampusFieldOpsHandler  campusFieldOps.Handler
}

func NewTenantModule(db *sql.DB) *TenantModule {
	tenantRepo := repository.NewTenantPostgresRepository(db)

	// Super Admin Services & Handlers
	createSvc := superTenantCreate.NewService(tenantRepo)
	listSvc := superTenantList.NewService(tenantRepo)
	assignBikeSvc := superBikeAssign.NewService(tenantRepo)

	// Campus Services & Handlers
	campusDashSvc := campusDashboard.NewService(tenantRepo)
	campusBikeListSvc := campusBikeList.NewService(tenantRepo)
	campusFieldOpsSvc := campusFieldOps.NewService(tenantRepo)

	return &TenantModule{
		CreateTenantHandler:    superTenantCreate.NewHandler(createSvc),
		ListTenantHandler:      superTenantList.NewHandler(listSvc),
		AssignBikeHandler:      superBikeAssign.NewHandler(assignBikeSvc),
		CampusDashboardHandler: campusDashboard.NewHandler(campusDashSvc),
		CampusBikeListHandler:  campusBikeList.NewHandler(campusBikeListSvc),
		CampusFieldOpsHandler:  campusFieldOps.NewHandler(campusFieldOpsSvc),
	}
}

