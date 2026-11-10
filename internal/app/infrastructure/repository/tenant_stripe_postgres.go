package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

func (r *TenantPostgresRepository) GetByStripeAccountID(ctx context.Context, stripeAccountID string) (*entity.Tenant, error) {
	row, err := r.queries.GetTenantByStripeAccountID(ctx, sql.NullString{String: stripeAccountID, Valid: stripeAccountID != ""})
	if err != nil {
		return nil, err
	}
	return mapTenantModelToEntity(row), nil
}

func (r *TenantPostgresRepository) UpdateStripeStatus(ctx context.Context, id uuid.UUID, stripeAccountID string, completed bool) error {
	return r.queries.UpdateTenantStripeAccount(ctx, sqlc.UpdateTenantStripeAccountParams{
		ID:                        id,
		StripeAccountID:           sql.NullString{String: stripeAccountID, Valid: stripeAccountID != ""},
		StripeOnboardingCompleted: completed,
	})
}

