package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Transaction, error)
}
