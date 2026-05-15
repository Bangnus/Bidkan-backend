package webhook

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type PaySolutionsWebhookPayload struct {
	OrderNo    string  `json:"order_no"`
	RefNo      string  `json:"ref_no"` // รหัสอ้างอิงของระบบ Payment
	Status     string  `json:"status"` // เช่น "CP" (Completed), "FA" (Failed)
	Amount     float64 `json:"amount"`
}

type Service interface {
	ProcessCallback(ctx context.Context, payload PaySolutionsWebhookPayload) error
}

type service struct {
	txRepo   repository.TransactionRepository
	userRepo repository.UserRepository
}

func NewService(txRepo repository.TransactionRepository, userRepo repository.UserRepository) Service {
	return &service{
		txRepo:   txRepo,
		userRepo: userRepo,
	}
}

func (s *service) ProcessCallback(ctx context.Context, payload PaySolutionsWebhookPayload) error {
	// 1. แปลง OrderNo กลับเป็น Transaction ID
	txID, err := uuid.Parse(payload.OrderNo)
	if err != nil {
		return errors.New("invalid order_no format")
	}

	// 2. ค้นหา Transaction
	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil || tx == nil {
		return errors.New("transaction not found")
	}

	// ถ้า Transaction นี้ถูกดำเนินการไปแล้ว ไม่ต้องทำซ้ำ
	if tx.Status == "completed" {
		return nil
	}

	// 3. เช็คสถานะการชำระเงินจาก PaySolutions (อิงตามสมมติฐานว่า CP = Completed)
	if payload.Status == "CP" { // Completed
		// อัปเดตสถานะ Transaction เป็นสำเร็จ
		err = s.txRepo.UpdateStatus(ctx, tx.ID, "completed", &payload.RefNo)
		if err != nil {
			return errors.New("failed to update transaction status")
		}

		// ดึงข้อมูล User เพื่อเอา Balance ปัจจุบัน
		user, err := s.userRepo.GetByID(ctx, tx.UserID)
		if err != nil {
			return errors.New("user not found")
		}

		currentBalance, _ := strconv.ParseFloat(user.WalletBalance, 64)
		newBalance := currentBalance + payload.Amount

		// บันทึกเงินเข้ากระเป๋า
		err = s.userRepo.UpdateBalance(ctx, user.ID, fmt.Sprintf("%.2f", newBalance))
		if err != nil {
			return errors.New("failed to update user balance")
		}

	} else {
		// จ่ายไม่สำเร็จ
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", &payload.RefNo)
		return errors.New("payment failed")
	}

	return nil
}
