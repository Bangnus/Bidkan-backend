package field_ops

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type SwapBatteryRequest struct {
	BikeID string `json:"bike_id"`
}

type Service struct {
	repo     domainRepo.TenantRepository
	compRepo domainRepo.StaffCompensationRepository
}

func NewService(repo domainRepo.TenantRepository, compRepo domainRepo.StaffCompensationRepository) *Service {
	return &Service{repo: repo, compRepo: compRepo}
}

func (s *Service) GetQueue(ctx context.Context, slug string) ([]*entity.BikeData, error) {
	tenant, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	bikes, err := s.repo.ListBikesByTenant(ctx, tenant.ID)
	if err != nil {
		return nil, err
	}

	queue := make([]*entity.BikeData, 0)
	for _, b := range bikes {
		if b.Battery < 20 {
			queue = append(queue, b)
		}
	}
	return queue, nil
}

func (s *Service) SwapBattery(ctx context.Context, slug string, req SwapBatteryRequest, staffUserID *uuid.UUID) error {
	if req.BikeID == "" {
		return errors.New("bike_id is required")
	}

	if err := s.repo.SwapBattery(ctx, req.BikeID); err != nil {
		return err
	}

	if s.compRepo != nil && staffUserID != nil && slug != "" {
		tenant, err := s.repo.GetBySlug(ctx, slug)
		if err == nil && tenant != nil {
			cfg, _ := s.compRepo.GetCompensationConfig(ctx, tenant.ID)
			earned := 0.0
			if cfg != nil && cfg.PayModel == "per_task" {
				earned = cfg.BatterySwapFee
			}
			_ = s.compRepo.LogStaffTask(ctx, &entity.StaffTaskLog{
				TenantID:     tenant.ID,
				StaffUserID:  *staffUserID,
				TaskType:     "battery_swap",
				TargetID:     req.BikeID,
				EarnedAmount: earned,
			})
		}
	}
	return nil
}
