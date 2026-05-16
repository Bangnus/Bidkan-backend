package repository

import (
	"context"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type CouponRepository interface {
	GetByCode(ctx context.Context, code string) (*entity.Coupon, error)
	CheckUsed(ctx context.Context, userID uuid.UUID, couponID uuid.UUID) (bool, error)
	Redeem(ctx context.Context, userID uuid.UUID, couponID uuid.UUID) error
}
