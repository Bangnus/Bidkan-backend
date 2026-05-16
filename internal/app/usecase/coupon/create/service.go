package create

import (
	"context"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
	"database/sql"
)

type Request struct {
	Code      string    `json:"code" validate:"required"`
	Type      string    `json:"type" validate:"required"` // credit, discount
	Value     string    `json:"value" validate:"required"`
	MinAmount string    `json:"min_amount"`
	StartAt   time.Time `json:"start_at"`
	MaxUses   int32     `json:"max_uses" validate:"required"`
	ExpiredAt time.Time `json:"expired_at" validate:"required"`
}

type Service interface {
	Execute(ctx context.Context, req Request) error
}

type service struct {
	queries *sqlc.Queries
}

func NewService(db *sql.DB) Service {
	return &service{
		queries: sqlc.New(db),
	}
}

func (s *service) Execute(ctx context.Context, req Request) error {
	minAmount := "0.00"
	if req.MinAmount != "" {
		minAmount = req.MinAmount
	}

	startAt := time.Now()
	if !req.StartAt.IsZero() {
		startAt = req.StartAt
	}

	return s.queries.CreateCoupon(ctx, sqlc.CreateCouponParams{
		ID:        uuid.New(),
		Code:      req.Code,
		Type:      req.Type,
		Value:     req.Value,
		MinAmount: sql.NullString{String: minAmount, Valid: true},
		StartAt:   startAt,
		MaxUses:   req.MaxUses,
		ExpiredAt: req.ExpiredAt,
	})
}
