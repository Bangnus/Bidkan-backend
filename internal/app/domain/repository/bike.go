package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type BikeRepository interface {
	Create(ctx context.Context, bike entity.BikeData) error
	GetByID(ctx context.Context, id string) (*entity.BikeData, error)
	ListAll(ctx context.Context) ([]entity.BikeData, error)
	ListAvailable(ctx context.Context) ([]entity.BikeData, error)
	UpdateStatus(ctx context.Context, data entity.BikeData) error
	UpdateRideState(ctx context.Context, bikeID, status string, currentRideID *uuid.UUID) error
	SaveLocation(ctx context.Context, data entity.BikeData) error
}
