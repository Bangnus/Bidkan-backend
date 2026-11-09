package field_ops

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type SwapBatteryRequest struct {
	BikeID string `json:"bike_id"`
}

type Service struct {
	repo domainRepo.TenantRepository
}

func NewService(repo domainRepo.TenantRepository) *Service {
	return &Service{repo: repo}
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

func (s *Service) SwapBattery(ctx context.Context, req SwapBatteryRequest) error {
	if req.BikeID == "" {
		return errors.New("bike_id is required")
	}
	return s.repo.SwapBattery(ctx, req.BikeID)
}

