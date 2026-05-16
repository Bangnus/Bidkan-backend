package start

import (
	"context"
	"errors"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Request struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	BikeID string    `json:"bike_id" validate:"required"`
}

type Service interface {
	Start(ctx context.Context, req Request) (*entity.Ride, error)
}

type service struct {
	rideRepo repository.RideRepository
	userRepo repository.UserRepository
	bikeRepo repository.BikeRepository
}

func NewService(rideRepo repository.RideRepository, userRepo repository.UserRepository, bikeRepo repository.BikeRepository) Service {
	return &service{
		rideRepo: rideRepo,
		userRepo: userRepo,
		bikeRepo: bikeRepo,
	}
}

func (s *service) Start(ctx context.Context, req Request) (*entity.Ride, error) {
	// 1. ตรวจสอบผู้ใช้
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	isStaff := user.Role == "staff" || user.Role == "admin"

	// 2. เช็คเงิน (ถ้าเป็น User ปกติ ต้องมีอย่างน้อย 20 บาท)
	if !isStaff {
		balance, _ := strconv.ParseFloat(user.WalletBalance, 64)
		if balance < 20.0 {
			return nil, errors.New("insufficient balance (minimum 20 THB required)")
		}
	}

	// 3. เช็คว่ามีการเช่าค้างอยู่หรือไม่
	active, _ := s.rideRepo.GetActiveRide(ctx, req.UserID)
	if active != nil {
		return nil, errors.New("you already have an ongoing ride")
	}

	// 4. เช็คสถานะรถจักรยาน
	bike, err := s.bikeRepo.GetByID(ctx, req.BikeID)
	if err != nil {
		return nil, errors.New("bike not found")
	}

	// ถ้าไม่ใช่พนักงาน ต้องเช็คสถานะรถและแบตเตอรี่
	if !isStaff {
		if bike.Status != "available" {
			return nil, errors.New("bike is not available")
		}
		if bike.Battery < 10 {
			return nil, errors.New("bike battery is too low")
		}
	}

	// 5. เริ่มการเช่า/ซ่อม
	rideType := "ride"
	if isStaff {
		rideType = "service"
	}

	rideID := uuid.New()
	ride := &entity.Ride{
		ID:       rideID,
		UserID:   req.UserID,
		BikeID:   req.BikeID,
		StartLat: bike.Lat,
		StartLon: bike.Lon,
		Status:   "ongoing",
		Type:     rideType,
	}

	if err := s.rideRepo.Create(ctx, ride); err != nil {
		return nil, errors.New("failed to start ride")
	}

	// 6. อัปเดตสถานะรถจักรยาน
	bike.Status = "in_use"
	rideIDStr := rideID.String()
	bike.CurrentRideID = &rideIDStr
	_ = s.bikeRepo.UpdateStatus(ctx, *bike) 

	return ride, nil
}
