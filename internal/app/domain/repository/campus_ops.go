package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type CampusOpsRepository interface {
	UpdateBikeStatus(ctx context.Context, bikeID string, status string) error
	ListZonesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantZone, error)
	CreateZoneForTenant(ctx context.Context, z *entity.TenantZone) error
	DeleteZone(ctx context.Context, zoneID uuid.UUID) error
	ListRidesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.CampusRideDetail, error)
	ListRidesByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.CampusRideDetail, int, error)
	GetTodayStats(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error)
	GetCompletedStats(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error)
	GetUnsettledStats(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error)
	GetUnsettledPeriod(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error)
	ListPayoutsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantPayout, error)
	ListPayoutsByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.TenantPayout, int, error)
	CompleteStripeOnboarding(ctx context.Context, tenantID uuid.UUID, stripeAccountID string) error
	ListStaffByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.CampusStaff, error)
	ListStaffByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.CampusStaff, int, error)
	ListAdminsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.CampusStaff, error)
	CreateStaffForTenant(ctx context.Context, u *entity.User, tenantID uuid.UUID) error
	CreateBackofficeUser(ctx context.Context, u *entity.User) error
	ListCouponsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Coupon, error)
	ListCouponsByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.Coupon, int, error)
	CreateCouponForTenant(ctx context.Context, c *entity.Coupon, tenantID uuid.UUID) error
	ListReportsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.MaintenanceReport, error)
	ResolveReport(ctx context.Context, reportID uuid.UUID, resolvedBy *uuid.UUID) error
}
