package repository

import (
	"context"
	"github.com/google/uuid"
)

type UserMonthlySpendingRepository interface {
	AddSpending(ctx context.Context, userID uuid.UUID, yearMonth int, amount float64) error
	GetTotalInWindow(ctx context.Context, userID uuid.UUID, sinceYearMonth int) (float64, error)
}
