package repository

import (
	"context"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
)

type ReportRepository interface {
	GetSummary(ctx context.Context) (*entity.FullReport, error)
}
