package repository

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"context"
)

type BikeRepository interface {
	SaveLocation(ctx context.Context, data entity.BikeData) error
	UpdateStatus(ctx context.Context, data entity.BikeData) error
}
