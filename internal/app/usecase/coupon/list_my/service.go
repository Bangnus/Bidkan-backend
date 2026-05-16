package list_my

import (
	"context"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
	"database/sql"
)

type CouponInfo struct {
	ID           uuid.UUID `json:"id"`
	Code         string    `json:"code"`
	Type         string    `json:"type"`
	Value        string    `json:"value"`
	MinAmount    string    `json:"min_amount"`
	ExpiredAt    time.Time `json:"expired_at"`
	CollectedAt  time.Time `json:"collected_at"`
	UserCouponID uuid.UUID `json:"user_coupon_id"`
}

type Service interface {
	Execute(ctx context.Context, userID uuid.UUID) ([]CouponInfo, error)
}

type service struct {
	queries *sqlc.Queries
}

func NewService(db *sql.DB) Service {
	return &service{
		queries: sqlc.New(db),
	}
}

func (s *service) Execute(ctx context.Context, userID uuid.UUID) ([]CouponInfo, error) {
	rows, err := s.queries.GetMyAvailableCoupons(ctx, userID)
	if err != nil {
		return nil, err
	}

	var coupons []CouponInfo
	for _, r := range rows {
		coupons = append(coupons, CouponInfo{
			ID:           r.ID,
			Code:         r.Code,
			Type:         r.Type,
			Value:        r.Value,
			MinAmount:    r.MinAmount.String,
			ExpiredAt:    r.ExpiredAt,
			CollectedAt:  r.CollectedAt,
			UserCouponID: r.UserCouponID,
		})
	}

	return coupons, nil
}
