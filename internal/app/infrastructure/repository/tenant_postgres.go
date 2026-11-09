package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type TenantPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewTenantPostgresRepository(db *sql.DB) domainRepo.TenantRepository {
	return &TenantPostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *TenantPostgresRepository) Create(ctx context.Context, t *entity.Tenant) error {
	return r.queries.CreateTenant(ctx, sqlc.CreateTenantParams{
		ID:                 t.ID,
		Slug:               t.Slug,
		Name:               t.Name,
		PlatformFeePercent: fmt.Sprintf("%.2f", t.PlatformFeePercent),
		UnlockFee:          fmt.Sprintf("%.2f", t.UnlockFee),
		PerMinuteRate:      fmt.Sprintf("%.2f", t.PerMinuteRate),
		Status:             t.Status,
	})
}

func (r *TenantPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	row, err := r.queries.GetTenantByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapTenantModelToEntity(row), nil
}

func (r *TenantPostgresRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	row, err := r.queries.GetTenantBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return mapTenantModelToEntity(row), nil
}

func (r *TenantPostgresRepository) List(ctx context.Context) ([]*entity.Tenant, error) {
	rows, err := r.queries.ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.Tenant, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapTenantModelToEntity(row))
	}
	return res, nil
}

func (r *TenantPostgresRepository) UpdateRates(ctx context.Context, id uuid.UUID, unlock, perMin, gp float64) error {
	return r.queries.UpdateTenantRates(ctx, sqlc.UpdateTenantRatesParams{
		ID:                 id,
		UnlockFee:          fmt.Sprintf("%.2f", unlock),
		PerMinuteRate:      fmt.Sprintf("%.2f", perMin),
		PlatformFeePercent: fmt.Sprintf("%.2f", gp),
	})
}

func (r *TenantPostgresRepository) AssignBike(ctx context.Context, bikeID string, tenantID *uuid.UUID) error {
	var nullUUID uuid.NullUUID
	if tenantID != nil {
		nullUUID = uuid.NullUUID{UUID: *tenantID, Valid: true}
	}
	return r.queries.AssignBikeToTenant(ctx, sqlc.AssignBikeToTenantParams{
		ID:       bikeID,
		TenantID: nullUUID,
	})
}

func (r *TenantPostgresRepository) ListBikesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.BikeData, error) {
	rows, err := r.queries.ListBikesByTenant(ctx, uuid.NullUUID{UUID: tenantID, Valid: true})
	if err != nil {
		return nil, err
	}
	res := make([]*entity.BikeData, 0, len(rows))
	for _, b := range rows {
		var hb sql.NullTime
		if b.LastHeartbeat.Valid {
			hb = b.LastHeartbeat
		}
		var rideID *string
		if b.CurrentRideID.Valid {
			str := b.CurrentRideID.UUID.String()
			rideID = &str
		}
		img := ""
		if b.ImageUrl.Valid {
			img = b.ImageUrl.String
		}
		res = append(res, &entity.BikeData{
			BikeID:        b.ID,
			HardwareID:    b.HardwareID,
			Lat:           b.Lat,
			Lon:           b.Lon,
			Battery:       int(b.BatteryLevel),
			Status:        b.Status,
			ImageURL:      img,
			CurrentRideID: rideID,
			LastHeartbeat: hb.Time,
		})
	}
	return res, nil
}

func (r *TenantPostgresRepository) SwapBattery(ctx context.Context, bikeID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE bikes SET battery_level = 100, status = 'available', updated_at = NOW() WHERE id = $1", bikeID)
	return err
}

func mapTenantModelToEntity(row sqlc.Tenant) *entity.Tenant {
	gp, _ := strconv.ParseFloat(row.PlatformFeePercent, 64)
	unlock, _ := strconv.ParseFloat(row.UnlockFee, 64)
	perMin, _ := strconv.ParseFloat(row.PerMinuteRate, 64)
	var stripeID *string
	if row.StripeAccountID.Valid {
		stripeID = &row.StripeAccountID.String
	}
	return &entity.Tenant{
		ID:                        row.ID,
		Slug:                      row.Slug,
		Name:                      row.Name,
		StripeAccountID:           stripeID,
		StripeOnboardingCompleted: row.StripeOnboardingCompleted,
		PlatformFeePercent:        gp,
		UnlockFee:                 unlock,
		PerMinuteRate:             perMin,
		Status:                    row.Status,
		CreatedAt:                 row.CreatedAt,
		UpdatedAt:                 row.UpdatedAt,
	}
}

