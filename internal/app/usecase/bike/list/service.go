package list

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Service interface {
	GetAvailableBikes(ctx context.Context) ([]entity.BikeData, error)
	GetAllBikes(ctx context.Context) ([]entity.BikeData, error)
}

type service struct {
	bikeRepo repository.BikeRepository
}

func NewService(bikeRepo repository.BikeRepository) Service {
	return &service{bikeRepo: bikeRepo}
}

func (s *service) GetAvailableBikes(ctx context.Context) ([]entity.BikeData, error) {
	return s.bikeRepo.ListAvailable(ctx)
}

func (s *service) GetAllBikes(ctx context.Context) ([]entity.BikeData, error) {
	return s.bikeRepo.ListAll(ctx)
}
