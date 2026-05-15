package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type transactionPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewTransactionPostgresRepository(db *sql.DB) repository.TransactionRepository {
	return &transactionPostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *transactionPostgresRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	_, err := r.db.ExecContext(ctx, 
		"INSERT INTO transactions (id, user_id, amount, type, status, reference_id, gateway_ref, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())",
		tx.ID, tx.UserID, tx.Amount, tx.Type, tx.Status, tx.ReferenceID, tx.GatewayRef,
	)
	return err
}

func (r *transactionPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, user_id, amount, type, status, reference_id, gateway_ref, created_at, updated_at FROM transactions WHERE id = $1", id)
	var tx entity.Transaction
	var gatewayRef sql.NullString
	err := row.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.Status, &tx.ReferenceID, &gatewayRef, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if gatewayRef.Valid {
		tx.GatewayRef = &gatewayRef.String
	}
	return &tx, nil
}

func (r *transactionPostgresRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, user_id, amount, type, status, reference_id, gateway_ref, created_at, updated_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []entity.Transaction
	for rows.Next() {
		var tx entity.Transaction
		var gatewayRef sql.NullString
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.Status, &tx.ReferenceID, &gatewayRef, &tx.CreatedAt, &tx.UpdatedAt); err == nil {
			if gatewayRef.Valid {
				tx.GatewayRef = &gatewayRef.String
			}
			res = append(res, tx)
		}
	}
	return res, nil
}

func (r *transactionPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, gatewayRef *string) error {
	_, err := r.db.ExecContext(ctx, 
		"UPDATE transactions SET status = $2, gateway_ref = COALESCE($3, gateway_ref), updated_at = NOW() WHERE id = $1", 
		id, status, gatewayRef,
	)
	return err
}

func (r *transactionPostgresRepository) GetTotalSpendingInWindow(ctx context.Context, userID uuid.UUID, since time.Time) (float64, error) {
	var totalStr string
	err := r.db.QueryRowContext(ctx, 
		"SELECT COALESCE(SUM(ABS(amount)), 0)::TEXT FROM transactions WHERE user_id = $1 AND type = 'fare_deduction' AND status = 'completed' AND created_at >= $2",
		userID, since,
	).Scan(&totalStr)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(totalStr, 64)
}

