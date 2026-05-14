package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	db "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
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
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (r *userPostgresRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err // sql.ErrNoRows ถ้าไม่เจอ
	}

	// Map จาก sqlc model → domain entity
	return &entity.User{
		ID:        row.ID,
		Email:     row.Email,
		Name:      row.Name,
		Password:  row.Password,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}