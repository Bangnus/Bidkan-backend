package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type SparePartRepository interface {
	ListByTenant(ctx context.Context, tenantID uuid.UUID, category string) ([]*entity.SparePart, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.SparePart, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, partCode string) (*entity.SparePart, error)
	Create(ctx context.Context, part *entity.SparePart) error
	AdjustStock(ctx context.Context, tenantID, partID uuid.UUID, delta int, changeType, refID, notes, actor string) (*entity.SparePart, error)
	ListUsageLogs(ctx context.Context, tenantID uuid.UUID, partID *uuid.UUID, limit int) ([]*entity.PartUsageLog, error)
}
