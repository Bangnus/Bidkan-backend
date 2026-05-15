package repository

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type userMonthlySpendingPostgresRepository struct {
	db *sql.DB
}

func NewUserMonthlySpendingRepository(db *sql.DB) repository.UserMonthlySpendingRepository {
	return &userMonthlySpendingPostgresRepository{db: db}
}

func (r *userMonthlySpendingPostgresRepository) AddSpending(ctx context.Context, userID uuid.UUID, yearMonth int, amount float64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO user_monthly_spending (user_id, year_month, amount, updated_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT (user_id, year_month) DO UPDATE SET amount = user_monthly_spending.amount + EXCLUDED.amount, updated_at = NOW()",
		userID, yearMonth, amount,
	)
	return err
}

func (r *userMonthlySpendingPostgresRepository) GetTotalInWindow(ctx context.Context, userID uuid.UUID, sinceYearMonth int) (float64, error) {
	var totalStr string
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount), 0)::TEXT FROM user_monthly_spending WHERE user_id = $1 AND year_month >= $2",
		userID, sinceYearMonth,
	).Scan(&totalStr)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(totalStr, 64)
}
