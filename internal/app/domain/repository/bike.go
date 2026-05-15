package repository

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"context"
)

type BikeRepository interface {
	Create(ctx context.Context, bike entity.BikeData) error
	GetByID(ctx context.Context, id string) (*entity.BikeData, error)
	ListAll(ctx context.Context) ([]entity.BikeData, error)
	ListAvailable(ctx context.Context) ([]entity.BikeData, error)
	UpdateStatus(ctx context.Context, data entity.BikeData) error
	SaveLocation(ctx context.Context, data entity.BikeData) error
}
