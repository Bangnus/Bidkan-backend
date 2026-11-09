package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"

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

func (r *bikePostgresRepository) Create(ctx context.Context, bike entity.BikeData) error {
	return r.queries.CreateBike(ctx, sqlc.CreateBikeParams{
		ID:         bike.BikeID,
		HardwareID: bike.HardwareID,
		Status:     bike.Status,
		ImageUrl:   sql.NullString{String: bike.ImageURL, Valid: bike.ImageURL != ""},
		TenantID:   toNullUUID(bike.TenantID),
	})
}

func (r *bikePostgresRepository) GetByID(ctx context.Context, id string) (*entity.BikeData, error) {
	b, err := r.queries.GetBike(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapBikeModelToEntity(b), nil
}

func (r *bikePostgresRepository) ListAll(ctx context.Context) ([]entity.BikeData, error) {
	bikes, err := r.queries.ListAllBikes(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapBikes(bikes), nil
}

func (r *bikePostgresRepository) ListAvailable(ctx context.Context) ([]entity.BikeData, error) {
	bikes, err := r.queries.ListAvailableBikes(ctx)
	if err != nil {
		return nil, err
	}
	return r.mapBikes(bikes), nil
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

func (r *bikePostgresRepository) UpdateRideState(ctx context.Context, bikeID, status string, currentRideID *uuid.UUID) error {
	return r.queries.UpdateBikeRideStatus(ctx, sqlc.UpdateBikeRideStatusParams{
		ID:            bikeID,
		Status:        status,
		CurrentRideID: toNullUUID(currentRideID),
	})
}

func (r *bikePostgresRepository) mapBikes(bikes []sqlc.Bike) []entity.BikeData {
	res := make([]entity.BikeData, 0, len(bikes))
	for _, b := range bikes {
		res = append(res, *mapBikeModelToEntity(b))
	}
	return res
}
