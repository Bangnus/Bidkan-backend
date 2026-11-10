package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

func (r *TenantPostgresRepository) ListPaged(ctx context.Context, limit, offset int) ([]*entity.Tenant, int, error) {
	total, err := r.queries.CountTenants(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListTenantsPaged(ctx, sqlc.ListTenantsPagedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.Tenant, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapTenantModelToEntity(row))
	}
	return res, int(total), nil
}

func (r *TenantPostgresRepository) ListBikesByTenantPaged(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*entity.BikeData, int, error) {
	if status == "all" {
		status = ""
	}
	tid := uuid.NullUUID{UUID: tenantID, Valid: true}
	total, err := r.queries.CountBikesByTenantFiltered(ctx, sqlc.CountBikesByTenantFilteredParams{
		TenantID: tid,
		Column2:  status,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListBikesByTenantPaged(ctx, sqlc.ListBikesByTenantPagedParams{
		TenantID: tid,
		Column2:  status,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	res := make([]*entity.BikeData, 0, len(rows))
	for _, b := range rows {
		res = append(res, mapBikeModelToEntity(b))
	}
	return res, int(total), nil
}
