package repository

import (
	"context"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	GetByGatewayRef(ctx context.Context, gatewayRef string) (*entity.Transaction, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, gatewayRef *string) error
	CompleteIfPending(ctx context.Context, id uuid.UUID, gatewayRef *string) (bool, error)
	GetTotalSpendingInWindow(ctx context.Context, userID uuid.UUID, since time.Time) (float64, error)
}
