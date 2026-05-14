package repository

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"context"
)

type BikeCacheRepository interface {
	SetLatestLocation(ctx context.Context, data entity.BikeData) error
	GetLatestLocation(ctx context.Context, bikeID string) (*entity.BikeData, error)
}
