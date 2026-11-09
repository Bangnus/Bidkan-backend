package repository

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

func (r *CampusOpsPostgresRepository) GetTodayStats(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error) {
	row, err := r.queries.GetTenantTodayStats(ctx, tenantID)
	if err != nil {
		return &entity.RideStatsSummary{}, err
	}
	gross, _ := strconv.ParseFloat(row.GrossVolume, 64)
	gp, _ := strconv.ParseFloat(row.GpFee, 64)
	net, _ := strconv.ParseFloat(row.NetEarnings, 64)
	return &entity.RideStatsSummary{
		TotalRides:  int(row.TotalTrips),
		GrossAmount: gross,
		GPFeeAmount: gp,
		NetAmount:   net,
	}, nil
}

func (r *CampusOpsPostgresRepository) GetCompletedStats(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error) {
	row, err := r.queries.GetTenantCompletedStats(ctx, tenantID)
	if err != nil {
		return &entity.RideStatsSummary{}, err
	}
	gross, _ := strconv.ParseFloat(row.GrossAmount, 64)
	gp, _ := strconv.ParseFloat(row.PlatformFeeAmount, 64)
	net, _ := strconv.ParseFloat(row.NetPayoutAmount, 64)
	return &entity.RideStatsSummary{
		TotalRides:  int(row.TotalRides),
		GrossAmount: gross,
		GPFeeAmount: gp,
		NetAmount:   net,
	}, nil
}

func (r *CampusOpsPostgresRepository) GetUnsettledStats(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error) {
	row, err := r.queries.GetTenantUnsettledStats(ctx, tenantID)
	if err != nil {
		return &entity.RideStatsSummary{}, err
	}
	gross, _ := strconv.ParseFloat(row.UnsettledGross, 64)
	gp, _ := strconv.ParseFloat(row.UnsettledGpFee, 64)
	net, _ := strconv.ParseFloat(row.UnsettledNet, 64)
	return &entity.RideStatsSummary{
		TotalRides:  int(row.UnsettledRides),
		GrossAmount: gross,
		GPFeeAmount: gp,
		NetAmount:   net,
	}, nil
}

func (r *CampusOpsPostgresRepository) ListCouponsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Coupon, error) {
	rows, err := r.queries.ListCouponsByTenant(ctx, uuid.NullUUID{UUID: tenantID, Valid: true})
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

func (r *CampusOpsPostgresRepository) CreateCouponForTenant(ctx context.Context, c *entity.Coupon, tenantID uuid.UUID) error {
	return r.queries.CreateScopedCoupon(ctx, sqlc.CreateScopedCouponParams{
		ID:        c.ID,
		TenantID:  uuid.NullUUID{UUID: tenantID, Valid: true},
		Scope:     "campus",
		Code:      c.Code,
		Type:      c.Type,
		Value:     c.Value,
		MinAmount: sql.NullString{String: c.MinAmount, Valid: c.MinAmount != ""},
		StartAt:   c.StartAt,
		MaxUses:   c.MaxUses,
		ExpiredAt: c.ExpiredAt,
	})
}

func (r *CampusOpsPostgresRepository) ListReportsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.MaintenanceReport, error) {
	rows, err := r.queries.ListReportsByTenant(ctx, uuid.NullUUID{UUID: tenantID, Valid: true})
	if err != nil {
		return nil, err
	}
	res := make([]*entity.MaintenanceReport, 0, len(rows))
	for _, row := range rows {
		var resolvedBy *uuid.UUID
		if row.ResolvedBy.Valid {
			u := row.ResolvedBy.UUID
			resolvedBy = &u
		}
		res = append(res, &entity.MaintenanceReport{
			ID:         row.ID,
			BikeID:     row.BikeID,
			ReportedBy: row.ReportedBy,
			IssueType:  row.IssueType,
			Status:     row.Status,
			ResolvedBy: resolvedBy,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		})
	}
	return res, nil
}

func (r *CampusOpsPostgresRepository) ResolveReport(ctx context.Context, reportID uuid.UUID, resolvedBy *uuid.UUID) error {
	return r.queries.ResolveMaintenanceReport(ctx, sqlc.ResolveMaintenanceReportParams{
		ID:         reportID,
		ResolvedBy: toNullUUID(resolvedBy),
	})
}
