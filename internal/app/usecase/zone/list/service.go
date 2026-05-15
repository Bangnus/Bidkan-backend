package list

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Service interface {
	GetAllZones(ctx context.Context) ([]entity.Zone, error)
}

type service struct {
	zoneRepo repository.ZoneRepository
}

func NewService(zoneRepo repository.ZoneRepository) Service {
	return &service{zoneRepo: zoneRepo}
}

func (s *service) GetAllZones(ctx context.Context) ([]entity.Zone, error) {
	return s.zoneRepo.ListAll(ctx)
}
