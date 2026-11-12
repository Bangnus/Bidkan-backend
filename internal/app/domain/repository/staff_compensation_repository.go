package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type StaffCompensationRepository interface {
	GetCompensationConfig(ctx context.Context, tenantID uuid.UUID) (*entity.StaffCompensationConfig, error)
	UpsertCompensationConfig(ctx context.Context, config *entity.StaffCompensationConfig) error
	LogStaffTask(ctx context.Context, log *entity.StaffTaskLog) error
	ListStaffEarnings(ctx context.Context, tenantID uuid.UUID) ([]*entity.StaffEarningSummary, error)
	GetStaffEarningsByUserID(ctx context.Context, tenantID, staffUserID uuid.UUID) (*entity.MyEarningsSummary, error)
}

