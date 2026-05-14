package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type ReportRepository interface {
	Create(ctx context.Context, report *entity.Report) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Report, error)
	ListByBike(ctx context.Context, bikeID string) ([]entity.Report, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, resolvedBy uuid.UUID) error
}
