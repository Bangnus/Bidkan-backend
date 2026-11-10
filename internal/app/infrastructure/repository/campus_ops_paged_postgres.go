package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

func (r *CampusOpsPostgresRepository) ListRidesByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.CampusRideDetail, int, error) {
	total, err := r.queries.CountRidesByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListRidesByTenantWithUserPaged(ctx, sqlc.ListRidesByTenantWithUserPagedParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.CampusRideDetail, 0, len(rows))
	for _, row := range rows {
		fare, _ := strconv.ParseFloat(row.TotalFare, 64)
		gp, _ := strconv.ParseFloat(row.PlatformFee, 64)
		net, _ := strconv.ParseFloat(row.PartnerEarnings, 64)
		var end *time.Time
		if row.EndTime.Valid {
			t := row.EndTime.Time
			end = &t
		}
		res = append(res, &entity.CampusRideDetail{
			ID:              row.ID,
			BikeID:          row.BikeID,
			UserID:          row.UserID,
			Username:        row.Username,
			PhoneNumber:     row.PhoneNumber.String,
			DistanceKm:      row.DistanceKm,
			TotalFare:       fare,
			PlatformFee:     gp,
			PartnerEarnings: net,
			Status:          row.Status,
			StartTime:       row.StartTime,
			EndTime:         end,
		})
	}
	return res, int(total), nil
}

func (r *CampusOpsPostgresRepository) ListPayoutsByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.TenantPayout, int, error) {
	total, err := r.queries.CountPayoutsByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListPayoutsByTenantPaged(ctx, sqlc.ListPayoutsByTenantPagedParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
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

func (r *CampusOpsPostgresRepository) ListStaffByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.CampusStaff, int, error) {
	tid := uuid.NullUUID{UUID: tenantID, Valid: true}
	total, err := r.queries.CountStaffByTenant(ctx, tid)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListStaffByTenantPaged(ctx, sqlc.ListStaffByTenantPagedParams{
		TenantID: tid,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	return mapCampusStaffRows(rows), int(total), nil
}

func (r *CampusOpsPostgresRepository) ListCouponsByTenantPaged(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.Coupon, int, error) {
	tid := uuid.NullUUID{UUID: tenantID, Valid: true}
	total, err := r.queries.CountCouponsByTenant(ctx, tid)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListCouponsByTenantPaged(ctx, sqlc.ListCouponsByTenantPagedParams{
		TenantID: tid,
		Limit:    int32(limit),
		Offset:   int32(offset),
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

