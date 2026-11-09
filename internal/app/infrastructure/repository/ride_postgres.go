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
	unlock := ride.UnlockFee
	if unlock == "" {
		unlock = "0.00"
	}
	perMin := ride.RatePerMinute
	if perMin == "" {
		perMin = "0.00"
	}
	return r.queries.CreateRide(ctx, sqlc.CreateRideParams{
		ID:            ride.ID,
		UserID:        ride.UserID,
		BikeID:        ride.BikeID,
		TenantID:      ride.TenantID,
		StartLat:      ride.StartLat,
		StartLon:      ride.StartLon,
		UnlockFee:     unlock,
		RatePerMinute: perMin,
		Type:          ride.Type,
	})
}

func (r *ridePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Ride, error) {
	res, err := r.queries.GetRide(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapRideModel(res), nil
}

func (r *ridePostgresRepository) EndRide(
	ctx context.Context,
	id uuid.UUID,
	endLat, endLon, distance float64,
	fare, platformFee, partnerEarnings string,
) error {
	if platformFee == "" {
		platformFee = "0.00"
	}
	if partnerEarnings == "" {
		partnerEarnings = "0.00"
	}
	return r.queries.EndRide(ctx, sqlc.EndRideParams{
		ID:              id,
		EndLat:          sql.NullFloat64{Float64: endLat, Valid: true},
		EndLon:          sql.NullFloat64{Float64: endLon, Valid: true},
		DistanceKm:      distance,
		TotalFare:       fare,
		PlatformFee:     platformFee,
		PartnerEarnings: partnerEarnings,
	})
}

func (r *ridePostgresRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Ride, error) {
	rides, err := r.queries.ListRidesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]entity.Ride, 0, len(rides))
	for _, row := range rides {
		res = append(res, *mapRideModel(row))
	}
	return res, nil
}

func (r *ridePostgresRepository) GetActiveRide(ctx context.Context, userID uuid.UUID) (*entity.Ride, error) {
	res, err := r.queries.GetActiveRideByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapRideModel(res), nil
}

func mapRideModel(res sqlc.Ride) *entity.Ride {
	ride := &entity.Ride{
		ID:              res.ID,
		UserID:          res.UserID,
		BikeID:          res.BikeID,
		TenantID:        res.TenantID,
		StartTime:       res.StartTime,
		StartLat:        res.StartLat,
		StartLon:        res.StartLon,
		DistanceKm:      res.DistanceKm,
		UnlockFee:       res.UnlockFee,
		RatePerMinute:   res.RatePerMinute,
		TotalFare:       res.TotalFare,
		PlatformFee:     res.PlatformFee,
		PartnerEarnings: res.PartnerEarnings,
		Status:          res.Status,
		Type:            res.Type,
		CreatedAt:       res.CreatedAt,
		UpdatedAt:       res.UpdatedAt,
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
	return ride
}
