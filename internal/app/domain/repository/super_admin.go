package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type SuperAdminRepository interface {
	ListAllBikesWithTenant(ctx context.Context) ([]*entity.BikeWithTenant, error)
	CreateBikeWithTenant(ctx context.Context, bikeID, hardwareID string, tenantID *uuid.UUID) error
	UpdateTenantStatus(ctx context.Context, id uuid.UUID, status string) error
	ListAllPayouts(ctx context.Context) ([]*entity.TenantPayout, error)
	CreatePayout(ctx context.Context, p *entity.TenantPayout) error
	ListGlobalCoupons(ctx context.Context) ([]*entity.Coupon, error)
	CreateGlobalCoupon(ctx context.Context, c *entity.Coupon) error
}
