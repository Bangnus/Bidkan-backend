package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type ridePostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewRidePostgresRepository(db *sql.DB) repository.RideRepository {
	return &ridePostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *ridePostgresRepository) Create(ctx context.Context, ride *entity.Ride) error {
	return r.queries.CreateRide(ctx, sqlc.CreateRideParams{
		ID:       ride.ID,
		UserID:   ride.UserID,
		BikeID:   ride.BikeID,
		StartLat: ride.StartLat,
		StartLon: ride.StartLon,
		Type:     ride.Type,
	})
}

func (r *ridePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Ride, error) {
	res, err := r.queries.GetRide(ctx, id)
	if err != nil {
		return nil, err
	}

	ride := &entity.Ride{
		ID:         res.ID,
		UserID:     res.UserID,
		BikeID:     res.BikeID,
		StartTime:  res.StartTime,
		StartLat:   res.StartLat,
		StartLon:   res.StartLon,
		DistanceKm: res.DistanceKm,
		TotalFare:  res.TotalFare,
		Status:     res.Status,
		Type:       res.Type,
		CreatedAt:  res.CreatedAt,
		UpdatedAt:  res.UpdatedAt,
	}

	if res.EndTime.Valid {
		ride.EndTime = &res.EndTime.Time
	}
	if res.EndLat.Valid {
		ride.EndLat = &res.EndLat.Float64
	}
	if res.EndLon.Valid {
		ride.EndLon = &res.EndLon.Float64
	}

	return ride, nil
}

func (r *ridePostgresRepository) EndRide(ctx context.Context, id uuid.UUID, endLat, endLon, distance float64, fare string) error {
	return r.queries.EndRide(ctx, sqlc.EndRideParams{
		ID:         id,
		EndLat:     sql.NullFloat64{Float64: endLat, Valid: true},
		EndLon:     sql.NullFloat64{Float64: endLon, Valid: true},
		DistanceKm: distance,
		TotalFare:  fare,
	})
}

func (r *ridePostgresRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Ride, error) {
	rides, err := r.queries.ListRidesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var res []entity.Ride
	for _, resRide := range rides {
		ride := entity.Ride{
			ID:         resRide.ID,
			UserID:     resRide.UserID,
			BikeID:     resRide.BikeID,
			StartTime:  resRide.StartTime,
			StartLat:   resRide.StartLat,
			StartLon:   resRide.StartLon,
			DistanceKm: resRide.DistanceKm,
			TotalFare:  resRide.TotalFare,
			Status:     resRide.Status,
			Type:       resRide.Type,
			CreatedAt:  resRide.CreatedAt,
			UpdatedAt:  resRide.UpdatedAt,
		}
		if resRide.EndTime.Valid {
			ride.EndTime = &resRide.EndTime.Time
		}
		if resRide.EndLat.Valid {
			ride.EndLat = &resRide.EndLat.Float64
		}
		if resRide.EndLon.Valid {
			ride.EndLon = &resRide.EndLon.Float64
		}
		res = append(res, ride)
	}
	return res, nil
}

func (r *ridePostgresRepository) GetActiveRide(ctx context.Context, userID uuid.UUID) (*entity.Ride, error) {
	res, err := r.queries.GetActiveRideByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	ride := &entity.Ride{
		ID:         res.ID,
		UserID:     res.UserID,
		BikeID:     res.BikeID,
		StartTime:  res.StartTime,
		StartLat:   res.StartLat,
		StartLon:   res.StartLon,
		DistanceKm: res.DistanceKm,
		TotalFare:  res.TotalFare,
		Status:     res.Status,
		Type:       res.Type,
		CreatedAt:  res.CreatedAt,
		UpdatedAt:  res.UpdatedAt,
	}

	if res.EndTime.Valid {
		ride.EndTime = &res.EndTime.Time
	}
	if res.EndLat.Valid {
		ride.EndLat = &res.EndLat.Float64
	}
	if res.EndLon.Valid {
		ride.EndLon = &res.EndLon.Float64
	}

	return ride, nil
}
