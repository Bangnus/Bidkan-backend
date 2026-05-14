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

func (r *bikePostgresRepository) Create(ctx context.Context, bikeID string, hardwareID string, status string) error {
	return r.queries.CreateBike(ctx, sqlc.CreateBikeParams{
		ID:         bikeID,
		HardwareID: hardwareID,
		Status:     status,
	})
}

func (r *bikePostgresRepository) GetByID(ctx context.Context, id string) (*entity.BikeData, error) {
	b, err := r.queries.GetBike(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.BikeData{
		BikeID:     b.ID,
		HardwareID: b.HardwareID,
		Lat:        b.Lat,
		Lon:        b.Lon,
		Battery:    int(b.BatteryLevel),
		Status:     b.Status,
	}, nil
}

func (r *bikePostgresRepository) ListAll(ctx context.Context) ([]entity.BikeData, error) {
	bikes, err := r.queries.ListAllBikes(ctx)
	if err != nil {
		return nil, err
	}
	var res []entity.BikeData
	for _, b := range bikes {
		res = append(res, entity.BikeData{
			BikeID:     b.ID,
			HardwareID: b.HardwareID,
			Lat:        b.Lat,
			Lon:        b.Lon,
			Battery:    int(b.BatteryLevel),
			Status:     b.Status,
		})
	}
	return res, nil
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
