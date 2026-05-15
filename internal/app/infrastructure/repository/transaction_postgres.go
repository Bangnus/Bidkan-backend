package repository

import (
	"context"
	"database/sql"

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
	return r.queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:          tx.ID,
		UserID:      tx.UserID,
		Amount:      tx.Amount,
		Type:        tx.Type,
		ReferenceID: tx.ReferenceID,
	})
}

func (r *transactionPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	res, err := r.queries.GetTransaction(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.Transaction{
		ID:          res.ID,
		UserID:      res.UserID,
		Amount:      res.Amount,
		Type:        res.Type,
		ReferenceID: res.ReferenceID,
		CreatedAt:   res.CreatedAt,
		UpdatedAt:   res.UpdatedAt,
	}, nil
}

func (r *transactionPostgresRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Transaction, error) {
	txs, err := r.queries.ListTransactionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var res []entity.Transaction
	for _, tx := range txs {
		res = append(res, entity.Transaction{
			ID:          tx.ID,
			UserID:      tx.UserID,
			Amount:      tx.Amount,
			Type:        tx.Type,
			ReferenceID: tx.ReferenceID,
			CreatedAt:   tx.CreatedAt,
			UpdatedAt:   tx.UpdatedAt,
		})
	}
	return res, nil
}
