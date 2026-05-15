package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateBalance(ctx context.Context, id uuid.UUID, newBalance string) error
	UpdateProfileImage(ctx context.Context, id uuid.UUID, imageURL string) error
}
