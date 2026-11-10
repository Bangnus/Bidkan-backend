package repository

import (
	"context"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

func (r *SuperAdminPostgresRepository) ListAllBikesWithTenantPaged(ctx context.Context, limit, offset int) ([]*entity.BikeWithTenant, int, error) {
	total, err := r.queries.CountAllBikes(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListAllBikesWithTenantPaged(ctx, sqlc.ListAllBikesWithTenantPagedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
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
	return res, int(total), nil
}

func (r *SuperAdminPostgresRepository) ListAllPayoutsPaged(ctx context.Context, limit, offset int) ([]*entity.TenantPayout, int, error) {
	total, err := r.queries.CountAllPayouts(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListAllPayoutsPaged(ctx, sqlc.ListAllPayoutsPagedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.TenantPayout, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapPayoutRow(row))
	}
	return res, int(total), nil
}

func (r *SuperAdminPostgresRepository) ListGlobalCouponsPaged(ctx context.Context, limit, offset int) ([]*entity.Coupon, int, error) {
	total, err := r.queries.CountGlobalCoupons(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListGlobalCouponsPaged(ctx, sqlc.ListGlobalCouponsPagedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.Coupon, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapCouponRow(row))
	}
	return res, int(total), nil
}

func mapCouponRow(row sqlc.Coupon) *entity.Coupon {
	return &entity.Coupon{
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
	}
}

