package redeem

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type Request struct {
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code"`
}

type Service interface {
	Execute(ctx context.Context, req Request) error
}

type service struct {
	db         *sql.DB
	couponRepo repository.CouponRepository
	userRepo   repository.UserRepository
	mqttPub    mqtt.Publisher
	queries    *sqlc.Queries
}

func NewService(db *sql.DB, couponRepo repository.CouponRepository, userRepo repository.UserRepository, mqttPub mqtt.Publisher) Service {
	return &service{
		db:         db,
		couponRepo: couponRepo,
		userRepo:   userRepo,
		mqttPub:    mqttPub,
		queries:    sqlc.New(db),
	}
}

func (s *service) Execute(ctx context.Context, req Request) error {
	// 1. ค้นหาคูปองจากโค้ด
	coupon, err := s.couponRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return errors.New("invalid coupon code")
	}

	// 2. ตรวจสอบวันเริ่มใช้งาน
	if time.Now().Before(coupon.StartAt) {
		return fmt.Errorf("coupon will be available at %s", coupon.StartAt.Format(time.RFC3339))
	}

	// 3. ตรวจสอบวันหมดอายุ
	if time.Now().After(coupon.ExpiredAt) {
		return errors.New("coupon has expired")
	}

	// 3. ตรวจสอบจำนวนครั้งที่ใช้ได้ทั้งหมด
	if coupon.UsedCount >= coupon.MaxUses {
		return errors.New("coupon has reached maximum usage limit")
	}

	// 4. ตรวจสอบว่าผู้ใช้เคยใช้โค้ดนี้ไปแล้วหรือยัง
	isUsed, err := s.couponRepo.CheckUsed(ctx, req.UserID, coupon.ID)
	if err != nil {
		return err
	}
	if isUsed {
		return errors.New("you have already used this coupon")
	}

	// 5. ดำเนินการตามประเภทของคูปอง (ตัวอย่างนี้เน้น 'credit' รับเงินฟรี)
	if coupon.Type == "credit" {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		qtx := s.queries.WithTx(tx)

		// A. เพิ่มยอดเงินให้ผู้ใช้
		err = qtx.AddBalance(ctx, sqlc.AddBalanceParams{
			WalletBalance: coupon.Value,
			ID:            req.UserID,
		})
		if err != nil {
			return err
		}

		// B. บันทึกประวัติ Transaction
		err = qtx.CreateTransaction(ctx, sqlc.CreateTransactionParams{
			ID:         uuid.New(),
			UserID:     req.UserID,
			Amount:     "+" + coupon.Value,
			Type:       "coupon_redeem",
			Status:     "completed",
			ReferenceID: sql.NullString{String: coupon.ID.String(), Valid: true},
		})
		if err != nil {
			return err
		}

		// C. บันทึกการใช้งานคูปอง
		err = qtx.IncrementCouponUsedCount(ctx, coupon.ID)
		if err != nil {
			return err
		}

		err = qtx.RecordUserCoupon(ctx, sqlc.RecordUserCouponParams{
			ID:       uuid.New(),
			UserID:   req.UserID,
			CouponID: coupon.ID,
			IsUsed:   true,
			UsedAt:   sql.NullTime{Time: time.Now(), Valid: true},
		})
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		// [NEW] ส่งแจ้งเตือน
		if s.mqttPub != nil {
			s.mqttPub.Publish(fmt.Sprintf("bidkan/users/%s/notifications", req.UserID), map[string]interface{}{
				"type":    "coupon_redeemed_credit",
				"amount":  coupon.Value,
				"code":    coupon.Code,
				"message": "รับเงินฟรีเข้าวอลเล็ตสำเร็จ",
				"time":    time.Now().Format(time.RFC3339),
			})
		}
		return nil
	}

	// กรณีประเภท 'discount' (เก็บไว้ใช้หักตอนจบงาน)
	if coupon.Type == "discount" {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		qtx := s.queries.WithTx(tx)

		// A. บันทึกว่าผู้ใช้ได้รับคูปองแล้ว (แต่ยังไม่ได้ใช้ is_used = false)
		err = qtx.RecordUserCoupon(ctx, sqlc.RecordUserCouponParams{
			ID:       uuid.New(),
			UserID:   req.UserID,
			CouponID: coupon.ID,
			IsUsed:   false,
			UsedAt:   sql.NullTime{Valid: false},
		})
		if err != nil {
			return err
		}

		// B. เพิ่มจำนวนการใช้งานรวมของคูปอง
		err = qtx.IncrementCouponUsedCount(ctx, coupon.ID)
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		// [NEW] ส่งแจ้งเตือน
		if s.mqttPub != nil {
			s.mqttPub.Publish(fmt.Sprintf("bidkan/users/%s/notifications", req.UserID), map[string]interface{}{
				"type":    "coupon_redeemed_discount",
				"code":    coupon.Code,
				"value":   coupon.Value,
				"message": "เก็บโค้ดส่วนลดสำเร็จ",
				"time":    time.Now().Format(time.RFC3339),
			})
		}
		return nil
	}

	return errors.New("unsupported coupon type")
}
