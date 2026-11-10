package bike_list

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Service struct {
	repo domainRepo.TenantRepository
}

func NewService(repo domainRepo.TenantRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Execute(ctx context.Context, slug string) ([]*entity.BikeData, error) {
	tenant, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	return s.repo.ListBikesByTenant(ctx, tenant.ID)
}

func (s *Service) ExecutePaged(ctx context.Context, slug, status string, limit, offset int) ([]*entity.BikeData, int, error) {
	tenant, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, 0, errors.New("campus tenant not found")
	}

	return s.repo.ListBikesByTenantPaged(ctx, tenant.ID, status, limit, offset)
}

