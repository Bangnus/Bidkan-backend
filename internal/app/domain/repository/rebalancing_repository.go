package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type RebalancingRepository interface {
	GetHeatmapPoints(ctx context.Context, tenantID uuid.UUID, limit int) ([]*entity.HeatmapPoint, error)
}

