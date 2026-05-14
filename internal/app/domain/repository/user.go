package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)
}
