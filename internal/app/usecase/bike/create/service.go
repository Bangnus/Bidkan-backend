package create

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Request struct {
	BikeID     string `json:"bike_id" validate:"required"`
	HardwareID string `json:"hardware_id" validate:"required"`
	ImageURL   string `json:"image_url"`
}

type Service interface {
	Create(ctx context.Context, req Request) error
}

type service struct {
	bikeRepo repository.BikeRepository
}

func NewService(bikeRepo repository.BikeRepository) Service {
	return &service{bikeRepo: bikeRepo}
}

func (s *service) Create(ctx context.Context, req Request) error {
	// เช็คว่ามี ID นี้หรือยัง
	existing, _ := s.bikeRepo.GetByID(ctx, req.BikeID)
	if existing != nil {
		return errors.New("bike ID already exists")
	}

	bike := entity.BikeData{
		BikeID:     req.BikeID,
		HardwareID: req.HardwareID,
		ImageURL:   req.ImageURL,
		Status:     "available", // ค่าเริ่มต้น
	}

	return s.bikeRepo.Create(ctx, bike)
}
