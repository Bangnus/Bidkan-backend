package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type IncidentRepository interface {
	ListByTenant(ctx context.Context, tenantID uuid.UUID, status, severity string) ([]*entity.IncidentReport, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.IncidentReport, error)
	Create(ctx context.Context, incident *entity.IncidentReport) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, approvedAmt float64, claimNo, notes string) (*entity.IncidentReport, error)
}
