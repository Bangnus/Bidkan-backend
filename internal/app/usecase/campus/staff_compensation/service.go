package staff_compensation

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type UpdateConfigRequest struct {
	PayModel        string  `json:"pay_model"`
	BatterySwapFee  float64 `json:"battery_swap_fee"`
	RepairTaskFee   float64 `json:"repair_task_fee"`
	DailyBaseWage   float64 `json:"daily_base_wage"`
	MonthlyBaseWage float64 `json:"monthly_base_wage"`
}

type Service struct {
	tenantRepo domainRepo.TenantRepository
	compRepo   domainRepo.StaffCompensationRepository
}

func NewService(tenantRepo domainRepo.TenantRepository, compRepo domainRepo.StaffCompensationRepository) *Service {
	return &Service{tenantRepo: tenantRepo, compRepo: compRepo}
}

func (s *Service) GetConfig(ctx context.Context, slug string) (*entity.StaffCompensationConfig, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}
	return s.compRepo.GetCompensationConfig(ctx, tenant.ID)
}

func (s *Service) UpdateConfig(ctx context.Context, slug string, req UpdateConfigRequest) (*entity.StaffCompensationConfig, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	if req.PayModel != "per_task" && req.PayModel != "daily" && req.PayModel != "monthly" {
		req.PayModel = "per_task"
	}
	if req.BatterySwapFee < 0 {
		req.BatterySwapFee = 0
	}
	if req.RepairTaskFee < 0 {
		req.RepairTaskFee = 0
	}
	if req.DailyBaseWage < 0 {
		req.DailyBaseWage = 0
	}
	if req.MonthlyBaseWage < 0 {
		req.MonthlyBaseWage = 0
	}

	cfg := &entity.StaffCompensationConfig{
		TenantID:        tenant.ID,
		PayModel:        req.PayModel,
		BatterySwapFee:  req.BatterySwapFee,
		RepairTaskFee:   req.RepairTaskFee,
		DailyBaseWage:   req.DailyBaseWage,
		MonthlyBaseWage: req.MonthlyBaseWage,
	}

	if err := s.compRepo.UpsertCompensationConfig(ctx, cfg); err != nil {
		return nil, err
	}
	return s.compRepo.GetCompensationConfig(ctx, tenant.ID)
}

func (s *Service) ListStaffEarnings(ctx context.Context, slug string) ([]*entity.StaffEarningSummary, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}
	return s.compRepo.ListStaffEarnings(ctx, tenant.ID)
}

func (s *Service) GetMyEarnings(ctx context.Context, slug string, staffUserID uuid.UUID) (*entity.MyEarningsSummary, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}
	return s.compRepo.GetStaffEarningsByUserID(ctx, tenant.ID, staffUserID)
}
