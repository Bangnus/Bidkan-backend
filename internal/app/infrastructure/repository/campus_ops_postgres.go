package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type CampusOpsPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewCampusOpsPostgresRepository(db *sql.DB) domainRepo.CampusOpsRepository {
	return &CampusOpsPostgresRepository{db: db, queries: sqlc.New(db)}
}

func (r *CampusOpsPostgresRepository) UpdateBikeStatus(ctx context.Context, bikeID string, status string) error {
	return r.queries.UpdateBikeOperationalStatus(ctx, sqlc.UpdateBikeOperationalStatusParams{
		ID:     bikeID,
		Status: status,
	})
}

func (r *CampusOpsPostgresRepository) ListZonesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantZone, error) {
	rows, err := r.queries.ListZonesByTenant(ctx, uuid.NullUUID{UUID: tenantID, Valid: true})
	if err != nil {
		return nil, err
	}
	res := make([]*entity.TenantZone, 0, len(rows))
	for _, z := range rows {
		res = append(res, &entity.TenantZone{
			ID:        z.ID,
			TenantID:  z.TenantID.UUID,
			Name:      z.Name,
			Type:      z.Type,
			Boundary:  z.Boundary,
			Radius:    z.Radius,
			CreatedAt: z.CreatedAt,
			UpdatedAt: z.UpdatedAt,
		})
	}
	return res, nil
}

func (r *CampusOpsPostgresRepository) CreateZoneForTenant(ctx context.Context, z *entity.TenantZone) error {
	return r.queries.CreateZone(ctx, sqlc.CreateZoneParams{
		ID:       z.ID,
		TenantID: uuid.NullUUID{UUID: z.TenantID, Valid: true},
		Name:     z.Name,
		Type:     z.Type,
		Boundary: z.Boundary,
		Radius:   z.Radius,
	})
}

func (r *CampusOpsPostgresRepository) DeleteZone(ctx context.Context, zoneID uuid.UUID) error {
	return r.queries.DeleteZone(ctx, zoneID)
}

func (r *CampusOpsPostgresRepository) ListRidesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.CampusRideDetail, error) {
	rows, err := r.queries.ListRidesByTenantWithUser(ctx, tenantID)
	if err != nil {
		return nil, err
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
			PhoneNumber:     row.PhoneNumber,
			DistanceKm:      row.DistanceKm,
			TotalFare:       fare,
			PlatformFee:     gp,
			PartnerEarnings: net,
			Status:          row.Status,
			StartTime:       row.StartTime,
			EndTime:         end,
		})
	}
	return res, nil
}

func (r *CampusOpsPostgresRepository) ListPayoutsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantPayout, error) {
	rows, err := r.queries.ListPayoutsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.TenantPayout, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapPayoutRow(row))
	}
	return res, nil
}

func (r *CampusOpsPostgresRepository) CompleteStripeOnboarding(ctx context.Context, tenantID uuid.UUID, stripeAccountID string) error {
	return r.queries.UpdateTenantStripeAccount(ctx, sqlc.UpdateTenantStripeAccountParams{
		ID:                        tenantID,
		StripeAccountID:           sql.NullString{String: stripeAccountID, Valid: true},
		StripeOnboardingCompleted: true,
	})
}

func (r *CampusOpsPostgresRepository) ListStaffByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.CampusStaff, error) {
	rows, err := r.queries.ListStaffByTenant(ctx, uuid.NullUUID{UUID: tenantID, Valid: true})
	if err != nil {
		return nil, err
	}
	res := make([]*entity.CampusStaff, 0, len(rows))
	for _, u := range rows {
		res = append(res, &entity.CampusStaff{
			ID:          u.ID,
			Username:    u.Username,
			PhoneNumber: u.PhoneNumber,
			Role:        u.Role,
			Status:      u.Status,
			CreatedAt:   u.CreatedAt,
		})
	}
	return res, nil
}

func (r *CampusOpsPostgresRepository) CreateStaffForTenant(ctx context.Context, u *entity.User, tenantID uuid.UUID) error {
	return r.queries.CreateTenantUser(ctx, sqlc.CreateTenantUserParams{
		ID:            u.ID,
		TenantID:      uuid.NullUUID{UUID: tenantID, Valid: true},
		PhoneNumber:   u.PhoneNumber,
		Username:      u.Username,
		Password:      u.Password,
		WalletBalance: u.WalletBalance,
		Role:          u.Role,
		Status:        u.Status,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
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
	var nullUser uuid.NullUUID
	if resolvedBy != nil {
		nullUser = uuid.NullUUID{UUID: *resolvedBy, Valid: true}
	}
	return r.queries.ResolveMaintenanceReport(ctx, sqlc.ResolveMaintenanceReportParams{
		ID:         reportID,
		ResolvedBy: nullUser,
	})
}
