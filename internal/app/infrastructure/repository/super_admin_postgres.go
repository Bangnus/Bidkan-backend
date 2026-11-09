package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type SuperAdminPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSuperAdminPostgresRepository(db *sql.DB) domainRepo.SuperAdminRepository {
	return &SuperAdminPostgresRepository{db: db, queries: sqlc.New(db)}
}

func (r *SuperAdminPostgresRepository) ListAllBikesWithTenant(ctx context.Context) ([]*entity.BikeWithTenant, error) {
	rows, err := r.queries.ListAllBikesWithTenant(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.BikeWithTenant, 0, len(rows))
	for _, row := range rows {
		var tID *uuid.UUID
		if row.TenantID.Valid {
			u := row.TenantID.UUID
			tID = &u
		}
		var hb *time.Time
		if row.LastHeartbeat.Valid {
			t := row.LastHeartbeat.Time
			hb = &t
		}
		res = append(res, &entity.BikeWithTenant{
			BikeID:        row.ID,
			HardwareID:    row.HardwareID,
			TenantID:      tID,
			TenantSlug:    row.TenantSlug,
			TenantName:    row.TenantName,
			Status:        row.Status,
			Lat:           row.Lat,
			Lon:           row.Lon,
			Battery:       int(row.BatteryLevel),
			LastHeartbeat: hb,
			CreatedAt:     row.CreatedAt,
		})
	}
	return res, nil
}

func (r *SuperAdminPostgresRepository) CreateBikeWithTenant(ctx context.Context, bikeID, hardwareID string, tenantID *uuid.UUID) error {
	var nullTenant uuid.NullUUID
	if tenantID != nil {
		nullTenant = uuid.NullUUID{UUID: *tenantID, Valid: true}
	}
	return r.queries.CreateBike(ctx, sqlc.CreateBikeParams{
		ID:         bikeID,
		HardwareID: hardwareID,
		Status:     "available",
		TenantID:   nullTenant,
	})
}

func (r *SuperAdminPostgresRepository) UpdateTenantStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.queries.UpdateTenantStatus(ctx, sqlc.UpdateTenantStatusParams{
		ID:     id,
		Status: status,
	})
}

func (r *SuperAdminPostgresRepository) ListAllPayouts(ctx context.Context) ([]*entity.TenantPayout, error) {
	rows, err := r.queries.ListAllPayouts(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.TenantPayout, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapPayoutRow(row))
	}
	return res, nil
}

func (r *SuperAdminPostgresRepository) CreatePayout(ctx context.Context, p *entity.TenantPayout) error {
	var trID sql.NullString
	if p.StripeTransferID != nil {
		trID = sql.NullString{String: *p.StripeTransferID, Valid: true}
	}
	return r.queries.CreateTenantPayout(ctx, sqlc.CreateTenantPayoutParams{
		ID:                p.ID,
		TenantID:          p.TenantID,
		StripeTransferID:  trID,
		TotalRides:        int32(p.TotalRides),
		GrossAmount:       fmt.Sprintf("%.2f", p.GrossAmount),
		PlatformFeeAmount: fmt.Sprintf("%.2f", p.PlatformFeeAmount),
		NetPayoutAmount:   fmt.Sprintf("%.2f", p.NetPayoutAmount),
		Status:            p.Status,
		PeriodStart:       p.PeriodStart,
		PeriodEnd:         p.PeriodEnd,
	})
}

func (r *SuperAdminPostgresRepository) ListGlobalCoupons(ctx context.Context) ([]*entity.Coupon, error) {
	rows, err := r.queries.ListGlobalCoupons(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.Coupon, 0, len(rows))
	for _, row := range rows {
		res = append(res, &entity.Coupon{
			ID:        row.ID,
			Code:      row.Code,
			Type:      row.Type,
			Value:     row.Value,
			MinAmount: row.MinAmount.String,
			StartAt:   row.StartAt,
			MaxUses:   row.MaxUses,
			UsedCount: row.UsedCount,
			ExpiredAt: row.ExpiredAt,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return res, nil
}

func (r *SuperAdminPostgresRepository) CreateGlobalCoupon(ctx context.Context, c *entity.Coupon) error {
	return r.queries.CreateScopedCoupon(ctx, sqlc.CreateScopedCouponParams{
		ID:        c.ID,
		TenantID:  uuid.NullUUID{Valid: false},
		Scope:     "global",
		Code:      c.Code,
		Type:      c.Type,
		Value:     c.Value,
		MinAmount: sql.NullString{String: c.MinAmount, Valid: c.MinAmount != ""},
		StartAt:   c.StartAt,
		MaxUses:   c.MaxUses,
		ExpiredAt: c.ExpiredAt,
	})
}

func mapPayoutRow(row sqlc.TenantPayout) *entity.TenantPayout {
	gross, _ := strconv.ParseFloat(row.GrossAmount, 64)
	fee, _ := strconv.ParseFloat(row.PlatformFeeAmount, 64)
	net, _ := strconv.ParseFloat(row.NetPayoutAmount, 64)
	var trID, errMsg *string
	if row.StripeTransferID.Valid {
		trID = &row.StripeTransferID.String
	}
	if row.ErrorMessage.Valid {
		errMsg = &row.ErrorMessage.String
	}
	var proc *time.Time
	if row.ProcessedAt.Valid {
		t := row.ProcessedAt.Time
		proc = &t
	}
	return &entity.TenantPayout{
		ID:                row.ID,
		TenantID:          row.TenantID,
		StripeTransferID:  trID,
		TotalRides:        int(row.TotalRides),
		GrossAmount:       gross,
		PlatformFeeAmount: fee,
		NetPayoutAmount:   net,
		Status:            row.Status,
		ErrorMessage:      errMsg,
		PeriodStart:       row.PeriodStart,
		PeriodEnd:         row.PeriodEnd,
		ProcessedAt:       proc,
		CreatedAt:         row.CreatedAt,
	}
}
