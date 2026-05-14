package repository

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type bikePostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewBikePostgresRepository(db *sql.DB) repository.BikeRepository {
	return &bikePostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *bikePostgresRepository) SaveLocation(ctx context.Context, data entity.BikeData) error {
	return r.queries.CreateBikeLocation(ctx, sqlc.CreateBikeLocationParams{
		ID:        uuid.New(),
		BikeID:    data.BikeID,
		Lat:       data.Lat,
		Lon:       data.Lon,
		Battery:   int32(data.Battery),
		CreatedAt: time.Now(),
	})
}

func (r *bikePostgresRepository) UpdateStatus(ctx context.Context, data entity.BikeData) error {
	return r.queries.UpdateBikeStatus(ctx, sqlc.UpdateBikeStatusParams{
		ID:           data.BikeID,
		Lat:          data.Lat,
		Lon:          data.Lon,
		BatteryLevel: int32(data.Battery),
	})
}
