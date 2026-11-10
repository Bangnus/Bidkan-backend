package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type OrganizationRepository interface {
	Create(ctx context.Context, org *entity.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Organization, error)
	List(ctx context.Context) ([]*entity.Organization, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type HardwareOrderRepository interface {
	Create(ctx context.Context, order *entity.HardwareOrder) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.HardwareOrder, error)
	ListByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.HardwareOrder, int, error)
	ListAll(ctx context.Context) ([]*entity.HardwareOrder, error)
	ListAllPaged(ctx context.Context, limit, offset int) ([]*entity.HardwareOrder, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}
