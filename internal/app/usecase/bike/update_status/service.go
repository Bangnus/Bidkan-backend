package update_status

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Request struct {
	BikeID string `json:"bike_id"`
	Status string `json:"status"`
}

type Service interface {
	Execute(ctx context.Context, req Request) error
}

type service struct {
	bikeRepo repository.BikeRepository
}

func NewService(bikeRepo repository.BikeRepository) Service {
	return &service{bikeRepo: bikeRepo}
}

func (s *service) Execute(ctx context.Context, req Request) error {
	bike, err := s.bikeRepo.GetByID(ctx, req.BikeID)
	if err != nil {
		return errors.New("bike not found")
	}

	// สามารถเพิ่ม Validation สถานะที่อนุญาตได้ที่นี่ เช่น available, maintenance, broken
	bike.Status = req.Status
	return s.bikeRepo.UpdateStatus(ctx, *bike)
}
