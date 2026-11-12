package dispute

import (
	"context"
	"errors"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Request struct {
	Reason string `json:"reason"`
}

type Service struct {
	rideRepo domainRepo.RideRepository
}

func NewService(rideRepo domainRepo.RideRepository) *Service {
	return &Service{rideRepo: rideRepo}
}

func (s *Service) DisputeRide(ctx context.Context, rideID, userID uuid.UUID, reason string) error {
	if reason == "" {
		return errors.New("dispute reason is required")
	}
	return s.rideRepo.DisputePenalty(ctx, rideID, userID, reason)
}
