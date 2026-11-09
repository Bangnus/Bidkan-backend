package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	List(ctx context.Context) ([]*entity.Tenant, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entity.Tenant, error)
	UpdateRates(ctx context.Context, id uuid.UUID, unlockFee, perMinuteRate, platformFee float64) error
	UpdateRatesFull(ctx context.Context, id uuid.UUID, unlockFee, perMinuteRate, platformFee, parkingPenalty, simFee float64, autoTiered bool) error
	RecalculateTieredGP(ctx context.Context, tenantID uuid.UUID) (float64, error)
	CountFleetForTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	AssignOrganization(ctx context.Context, id uuid.UUID, orgID *uuid.UUID) error
	AssignBike(ctx context.Context, bikeID string, tenantID *uuid.UUID) error
	ListBikesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.BikeData, error)
	SwapBattery(ctx context.Context, bikeID string) error
}
