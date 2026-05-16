package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type couponPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewCouponPostgresRepository(db *sql.DB) repository.CouponRepository {
	return &couponPostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *couponPostgresRepository) GetByCode(ctx context.Context, code string) (*entity.Coupon, error) {
	c, err := r.queries.GetCouponByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return &entity.Coupon{
		ID:        c.ID,
		Code:      c.Code,
		Type:      c.Type,
		Value:     c.Value,
		MinAmount: c.MinAmount.String,
		MaxUses:   c.MaxUses,
		UsedCount: c.UsedCount,
		ExpiredAt: c.ExpiredAt,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}, nil
}

func (r *couponPostgresRepository) CheckUsed(ctx context.Context, userID uuid.UUID, couponID uuid.UUID) (bool, error) {
	return r.queries.CheckUserCouponUsed(ctx, sqlc.CheckUserCouponUsedParams{
		UserID:   userID,
		CouponID: couponID,
	})
}

func (r *couponPostgresRepository) Redeem(ctx context.Context, userID uuid.UUID, couponID uuid.UUID) error {
	// รันแบบ Atomic (สำหรับ Increment และ Record การใช้)
	err := r.queries.IncrementCouponUsedCount(ctx, couponID)
	if err != nil {
		return err
	}

	return r.queries.RecordUserCoupon(ctx, sqlc.RecordUserCouponParams{
		ID:       uuid.New(),
		UserID:   userID,
		CouponID: couponID,
	})
}
