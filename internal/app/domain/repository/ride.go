package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type RideRepository interface {
	Create(ctx context.Context, ride *entity.Ride) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Ride, error)
	EndRide(ctx context.Context, id uuid.UUID, endLat, endLon, distance float64, fare string) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Ride, error)
	GetActiveRide(ctx context.Context, userID uuid.UUID) (*entity.Ride, error)
}
