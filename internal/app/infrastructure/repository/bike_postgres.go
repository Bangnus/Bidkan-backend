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
	})
}

func (r *bikePostgresRepository) GetByID(ctx context.Context, id string) (*entity.BikeData, error) {
	b, err := r.queries.GetBike(ctx, id)
	if err != nil {
		return nil, err
	}

	var rideID *string
	if b.CurrentRideID.Valid {
		str := b.CurrentRideID.UUID.String()
		rideID = &str
	}

	return &entity.BikeData{
		BikeID:        b.ID,
		HardwareID:    b.HardwareID,
		Lat:           b.Lat,
		Lon:           b.Lon,
		Battery:       int(b.BatteryLevel),
		Status:        b.Status,
		ImageURL:      b.ImageUrl.String,
		CurrentRideID: rideID,
		LastHeartbeat: b.LastHeartbeat,
	}, nil
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

// Helper function to map sqlc results to entities
func (r *bikePostgresRepository) mapBikes(bikes []sqlc.Bike) []entity.BikeData {
	var res []entity.BikeData
	for _, b := range bikes {
		var rideID *string
		if b.CurrentRideID.Valid {
			str := b.CurrentRideID.UUID.String()
			rideID = &str
		}

		res = append(res, entity.BikeData{
			BikeID:        b.ID,
			HardwareID:    b.HardwareID,
			Lat:           b.Lat,
			Lon:           b.Lon,
			Battery:       int(b.BatteryLevel),
			Status:        b.Status,
			ImageURL:      b.ImageUrl.String,
			CurrentRideID: rideID,
			LastHeartbeat: b.LastHeartbeat,
		})
	}
	return res
}
