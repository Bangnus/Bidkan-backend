package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	db "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type userPostgresRepository struct {
	queries *db.Queries
}

// คืนค่าเป็น Interface เพื่อให้ Service เรียกใช้ได้โดยไม่รู้ว่าเป็น DB อะไร
func NewUserPostgresRepository(conn *sql.DB) repository.UserRepository {
	return &userPostgresRepository{
		queries: db.New(conn),
	}
}

func (r *userPostgresRepository) Create(ctx context.Context, user *entity.User) error {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		ID:            user.ID,
		PhoneNumber:   user.PhoneNumber,
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
	row, err := r.queries.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, err // sql.ErrNoRows ถ้าไม่เจอ
	}

	// Map จาก sqlc model → domain entity
	return &entity.User{
		ID:            row.ID,
		PhoneNumber:   row.PhoneNumber,
		Username:      row.Username,
		Password:      row.Password,
		WalletBalance: row.WalletBalance,
		Role:          row.Role,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}

func (r *userPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.queries.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
		ID:     id,
		Status: status,
	})
}