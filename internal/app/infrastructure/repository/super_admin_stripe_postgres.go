package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

func (r *SuperAdminPostgresRepository) GetPayoutByID(ctx context.Context, id uuid.UUID) (*entity.TenantPayout, error) {
	row, err := r.queries.GetTenantPayoutByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapPayoutRow(row), nil
}

func (r *SuperAdminPostgresRepository) GetPayoutByTransferID(ctx context.Context, transferID string) (*entity.TenantPayout, error) {
	row, err := r.queries.GetTenantPayoutByTransferID(ctx, sql.NullString{String: transferID, Valid: transferID != ""})
	if err != nil {
		return nil, err
	}
	return mapPayoutRow(row), nil
}

func (r *SuperAdminPostgresRepository) HasInFlightPayout(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	return r.queries.HasInFlightPayoutForTenant(ctx, tenantID)
}

func (r *SuperAdminPostgresRepository) UpdatePayoutStatus(ctx context.Context, id uuid.UUID, status string, transferID *string, errMsg *string) error {
	var trID sql.NullString
	if transferID != nil && *transferID != "" {
		trID = sql.NullString{String: *transferID, Valid: true}
	}
	var errVal sql.NullString
	if errMsg != nil && *errMsg != "" {
		errVal = sql.NullString{String: *errMsg, Valid: true}
	}
	return r.queries.UpdatePayoutStatus(ctx, sqlc.UpdatePayoutStatusParams{
		ID:               id,
		Status:           status,
		StripeTransferID: trID,
		ErrorMessage:     errVal,
	})
}

func (r *SuperAdminPostgresRepository) TryRecordStripeEvent(ctx context.Context, eventID, eventType string) (bool, error) {
	rows, err := r.queries.TryRecordStripeEvent(ctx, sqlc.TryRecordStripeEventParams{
		EventID:   eventID,
		EventType: eventType,
	})
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

