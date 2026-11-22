package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

func (r *userPostgresRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}

func (r *userPostgresRepository) CreateEmailUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	var tID uuid.NullUUID
	if user.TenantID != nil {
		tID = uuid.NullUUID{UUID: *user.TenantID, Valid: true}
	}
	u, err := r.queries.CreateEmailUser(ctx, sqlc.CreateEmailUserParams{
		ID:          user.ID,
		TenantID:    tID,
		Email:       sql.NullString{String: user.Email, Valid: user.Email != ""},
		Username:    user.Username,
		PhoneNumber: sql.NullString{String: user.PhoneNumber, Valid: user.PhoneNumber != ""},
		Password:    user.Password,
	})
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}

func (r *userPostgresRepository) RefreshPendingEmailUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	var tID uuid.NullUUID
	if user.TenantID != nil {
		tID = uuid.NullUUID{UUID: *user.TenantID, Valid: true}
	}
	u, err := r.queries.RefreshPendingEmailUser(ctx, sqlc.RefreshPendingEmailUserParams{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: sql.NullString{String: user.PhoneNumber, Valid: user.PhoneNumber != ""},
		Password:    user.Password,
		TenantID:    tID,
	})
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}

func (r *userPostgresRepository) ActivateVerifiedEmailUser(ctx context.Context, email string) (*entity.User, error) {
	u, err := r.queries.ActivateVerifiedEmailUser(ctx, email)
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}
