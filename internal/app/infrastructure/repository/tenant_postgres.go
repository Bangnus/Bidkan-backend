package repository

import (
	"context"
	"database/sql"
	"fmt"

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
	return &TenantPostgresRepository{db: db, queries: sqlc.New(db)}
}

func (r *TenantPostgresRepository) Create(ctx context.Context, t *entity.Tenant) error {
	if t.ParkingPenaltyFee <= 0 {
		t.ParkingPenaltyFee = 50.0
	}
	return r.queries.CreateTenant(ctx, sqlc.CreateTenantParams{
		ID:                 t.ID,
		OrganizationID:     toNullUUID(t.OrganizationID),
		Slug:               t.Slug,
		Name:               t.Name,
		PlatformFeePercent: fmt.Sprintf("%.2f", t.PlatformFeePercent),
		AutoTieredGp:       t.AutoTieredGP,
		UnlockFee:          fmt.Sprintf("%.2f", t.UnlockFee),
		PerMinuteRate:      fmt.Sprintf("%.2f", t.PerMinuteRate),
		ParkingPenaltyFee:  fmt.Sprintf("%.2f", t.ParkingPenaltyFee),
		SimFeePerBike:      fmt.Sprintf("%.2f", t.SimFeePerBike),
		ContactName:        toNullString(t.ContactName),
		ContactEmail:       toNullString(t.ContactEmail),
		ContactPhone:       toNullString(t.ContactPhone),
		Notes:              toNullString(t.Notes),
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

func (r *TenantPostgresRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entity.Tenant, error) {
	rows, err := r.queries.ListTenantsByOrganization(ctx, uuid.NullUUID{UUID: orgID, Valid: true})
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
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return r.UpdateRatesFull(ctx, id, unlock, perMin, gp, existing.ParkingPenaltyFee, existing.SimFeePerBike, existing.AutoTieredGP)
}

func (r *TenantPostgresRepository) UpdateRatesFull(ctx context.Context, id uuid.UUID, unlock, perMin, gp, penalty, simFee float64, autoTiered bool) error {
	if penalty < 0 {
		penalty = 0
	}
	return r.queries.UpdateTenantRates(ctx, sqlc.UpdateTenantRatesParams{
		ID:                 id,
		UnlockFee:          fmt.Sprintf("%.2f", unlock),
		PerMinuteRate:      fmt.Sprintf("%.2f", perMin),
		PlatformFeePercent: fmt.Sprintf("%.2f", gp),
		AutoTieredGp:       autoTiered,
		ParkingPenaltyFee:  fmt.Sprintf("%.2f", penalty),
		SimFeePerBike:      fmt.Sprintf("%.2f", simFee),
	})
}

func (r *TenantPostgresRepository) CountFleetForTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	t, err := r.GetByID(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	if t.OrganizationID != nil {
		cnt, err := r.queries.CountBikesByOrganization(ctx, uuid.NullUUID{UUID: *t.OrganizationID, Valid: true})
		return int(cnt), err
	}
	cnt, err := r.queries.CountBikesByTenant(ctx, uuid.NullUUID{UUID: tenantID, Valid: true})
	return int(cnt), err
}

func (r *TenantPostgresRepository) RecalculateTieredGP(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	t, err := r.GetByID(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	if !t.AutoTieredGP {
		return t.PlatformFeePercent, nil
	}
	count, err := r.CountFleetForTenant(ctx, tenantID)
	if err != nil {
		return t.PlatformFeePercent, err
	}
	tier := entity.CalculateTieredGP(count)
	if t.OrganizationID != nil {
		peers, err := r.ListByOrganization(ctx, *t.OrganizationID)
		if err == nil {
			for _, p := range peers {
				if p.AutoTieredGP {
					_ = r.UpdateRatesFull(ctx, p.ID, p.UnlockFee, p.PerMinuteRate, tier.CurrentGPPercent, p.ParkingPenaltyFee, p.SimFeePerBike, true)
				}
			}
			return tier.CurrentGPPercent, nil
		}
	}
	if err := r.UpdateRatesFull(ctx, t.ID, t.UnlockFee, t.PerMinuteRate, tier.CurrentGPPercent, t.ParkingPenaltyFee, t.SimFeePerBike, true); err != nil {
		return t.PlatformFeePercent, err
	}
	return tier.CurrentGPPercent, nil
}

func (r *TenantPostgresRepository) AssignOrganization(ctx context.Context, id uuid.UUID, orgID *uuid.UUID) error {
	return r.queries.AssignTenantOrganization(ctx, sqlc.AssignTenantOrganizationParams{
		ID:             id,
		OrganizationID: toNullUUID(orgID),
	})
}

func (r *TenantPostgresRepository) AssignBike(ctx context.Context, bikeID string, tenantID *uuid.UUID) error {
	if err := r.queries.AssignBikeToTenant(ctx, sqlc.AssignBikeToTenantParams{
		ID:       bikeID,
		TenantID: toNullUUID(tenantID),
	}); err != nil {
		return err
	}
	if tenantID != nil {
		_, _ = r.RecalculateTieredGP(ctx, *tenantID)
	}
	return nil
}

func (r *TenantPostgresRepository) ListBikesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.BikeData, error) {
	rows, err := r.queries.ListBikesByTenant(ctx, uuid.NullUUID{UUID: tenantID, Valid: true})
	if err != nil {
		return nil, err
	}
	res := make([]*entity.BikeData, 0, len(rows))
	for _, b := range rows {
		res = append(res, mapBikeModelToEntity(b))
	}
	return res, nil
}

func (r *TenantPostgresRepository) SwapBattery(ctx context.Context, bikeID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE bikes SET battery_level = 100, status = 'available', updated_at = NOW() WHERE id = $1", bikeID)
	return err
}
