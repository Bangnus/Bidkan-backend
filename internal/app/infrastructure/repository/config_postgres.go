package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
)

type configPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewConfigRepository(db *sql.DB) repository.ConfigRepository {
	return &configPostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *configPostgresRepository) GetConfig(ctx context.Context, key string) (string, error) {
	val, err := r.queries.GetConfig(ctx, key)
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *configPostgresRepository) SetConfig(ctx context.Context, key string, value string) error {
	return r.queries.SetConfig(ctx, sqlc.SetConfigParams{
		Key:   key,
		Value: value,
	})
}
