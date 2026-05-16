package summary

import (
	"context"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Service interface {
	GetSummary(ctx context.Context) (*entity.FullReport, error)
}

type service struct {
	reportRepo repository.ReportRepository
}

func NewService(reportRepo repository.ReportRepository) Service {
	return &service{reportRepo: reportRepo}
}

func (s *service) GetSummary(ctx context.Context) (*entity.FullReport, error) {
	return s.reportRepo.GetSummary(ctx)
}
