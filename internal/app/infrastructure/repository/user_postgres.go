package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type userPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewUserPostgresRepository(db *sql.DB) repository.UserRepository {
	return &userPostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *userPostgresRepository) Create(ctx context.Context, user *entity.User) error {
	return r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:            user.ID,
		PhoneNumber:   sql.NullString{String: user.PhoneNumber, Valid: user.PhoneNumber != ""},
		Username:      user.Username,
		Password:      user.Password,
		WalletBalance: user.WalletBalance,
		Role:          user.Role,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	})
}

func (r *userPostgresRepository) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	u, err := r.queries.GetUserByPhone(ctx, sql.NullString{String: phone, Valid: phone != ""})
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}

func (r *userPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}

func (r *userPostgresRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	u, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}

func (r *userPostgresRepository) GetByIdentifier(ctx context.Context, identifier string) (*entity.User, error) {
	u, err := r.queries.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}

func (r *userPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.queries.UpdateUserStatus(ctx, sqlc.UpdateUserStatusParams{
		Status: status,
		ID:     id,
	})
}

func (r *userPostgresRepository) UpdateBalance(ctx context.Context, id uuid.UUID, newBalance string) error {
	return r.queries.UpdateUserBalance(ctx, sqlc.UpdateUserBalanceParams{
		WalletBalance: newBalance,
		ID:            id,
	})
}

func (r *userPostgresRepository) AddBalance(ctx context.Context, id uuid.UUID, amount string) error {
	return r.queries.AddBalance(ctx, sqlc.AddBalanceParams{
		WalletBalance: amount,
		ID:            id,
	})
}

func (r *userPostgresRepository) DeductBalance(ctx context.Context, id uuid.UUID, amount string) error {
	return r.queries.DeductBalance(ctx, sqlc.DeductBalanceParams{
		WalletBalance: amount,
		ID:            id,
	})
}

func (r *userPostgresRepository) UpdateProfileImage(ctx context.Context, id uuid.UUID, imageURL string) error {
	return r.queries.UpdateUserProfileImage(ctx, sqlc.UpdateUserProfileImageParams{
		ImageUrl: sql.NullString{String: imageURL, Valid: imageURL != ""},
		ID:       id,
	})
}

func (r *userPostgresRepository) UpsertGoogleUser(ctx context.Context, googleID, email, name, imageURL string, tenantID *uuid.UUID) (*entity.User, error) {
	var tID uuid.NullUUID
	if tenantID != nil {
		tID = uuid.NullUUID{UUID: *tenantID, Valid: true}
	}
	u, err := r.queries.UpsertGoogleUser(ctx, sqlc.UpsertGoogleUserParams{
		ID:          uuid.New(),
		TenantID:    tID,
		GoogleID:    sql.NullString{String: googleID, Valid: googleID != ""},
		Email:       sql.NullString{String: email, Valid: email != ""},
		Username:    name,
		PhoneNumber: sql.NullString{String: "", Valid: false},
		ImageUrl:    sql.NullString{String: imageURL, Valid: imageURL != ""},
	})
	if err != nil {
		return nil, err
	}
	return r.mapUser(u), nil
}

func (r *userPostgresRepository) mapUser(u sqlc.User) *entity.User {
	var tenantID *uuid.UUID
	if u.TenantID.Valid {
		tID := u.TenantID.UUID
		tenantID = &tID
	}
	return &entity.User{
		ID:            u.ID,
		TenantID:      tenantID,
		PhoneNumber:   u.PhoneNumber.String,
		Email:         u.Email.String,
		Username:      u.Username,
		Password:      u.Password,
		WalletBalance: u.WalletBalance,
		Role:          u.Role,
		Status:        u.Status,
		ImageURL:      u.ImageUrl.String,
		FcmToken:      u.FcmToken.String,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}