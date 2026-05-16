package entity

import (
	"time"
	"github.com/google/uuid"
)

type Coupon struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Type      string    `json:"type"` // credit, discount
	Value     string    `json:"value"`
	MinAmount string    `json:"min_amount"`
	StartAt   time.Time `json:"start_at"`
	MaxUses   int32     `json:"max_uses"`
	UsedCount int32     `json:"used_count"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserCoupon struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	CouponID uuid.UUID `json:"coupon_id"`
	UsedAt   time.Time `json:"used_at"`
}
