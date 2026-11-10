package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type SuperAdminRepository interface {
	ListAllBikesWithTenant(ctx context.Context) ([]*entity.BikeWithTenant, error)
	ListAllBikesWithTenantPaged(ctx context.Context, limit, offset int) ([]*entity.BikeWithTenant, int, error)
	CreateBikeWithTenant(ctx context.Context, bikeID, hardwareID string, tenantID *uuid.UUID) error
	UpdateTenantStatus(ctx context.Context, id uuid.UUID, status string) error
	ListAllPayouts(ctx context.Context) ([]*entity.TenantPayout, error)
	ListAllPayoutsPaged(ctx context.Context, limit, offset int) ([]*entity.TenantPayout, int, error)
	GetPayoutByID(ctx context.Context, id uuid.UUID) (*entity.TenantPayout, error)
	GetPayoutByTransferID(ctx context.Context, transferID string) (*entity.TenantPayout, error)
	HasInFlightPayout(ctx context.Context, tenantID uuid.UUID) (bool, error)
	CreatePayout(ctx context.Context, p *entity.TenantPayout) error
	UpdatePayoutStatus(ctx context.Context, id uuid.UUID, status string, transferID *string, errMsg *string) error
	TryRecordStripeEvent(ctx context.Context, eventID, eventType string) (bool, error)
	ListGlobalCoupons(ctx context.Context) ([]*entity.Coupon, error)
	ListGlobalCouponsPaged(ctx context.Context, limit, offset int) ([]*entity.Coupon, int, error)
	CreateGlobalCoupon(ctx context.Context, c *entity.Coupon) error
	ActivatePendingAdmins(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID) error
}
