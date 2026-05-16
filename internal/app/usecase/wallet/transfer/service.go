package transfer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type Request struct {
	SenderID      uuid.UUID `json:"sender_id"`
	ReceiverPhone string    `json:"receiver_phone"`
	Amount        float64   `json:"amount"`
}

type Service interface {
	Execute(ctx context.Context, req Request) error
}

type service struct {
	db        *sql.DB
	userRepo  repository.UserRepository
	txRepo    repository.TransactionRepository
	queries   *sqlc.Queries // สำหรับรันใน Transaction
}

func NewService(db *sql.DB, userRepo repository.UserRepository, txRepo repository.TransactionRepository) Service {
	return &service{
		db:       db,
		userRepo: userRepo,
		txRepo:   txRepo,
		queries:  sqlc.New(db),
	}
}

func (s *service) Execute(ctx context.Context, req Request) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	// 1. ตรวจสอบผู้โอน
	sender, err := s.userRepo.GetByID(ctx, req.SenderID)
	if err != nil {
		return errors.New("sender not found")
	}

	// 2. ตรวจสอบยอดเงินคงเหลือ
	var balance float64
	fmt.Sscanf(sender.WalletBalance, "%f", &balance)
	if balance < req.Amount {
		return errors.New("insufficient balance")
	}

	// 3. ตรวจสอบผู้รับจากเบอร์โทรศัพท์
	receiver, err := s.userRepo.GetByPhone(ctx, req.ReceiverPhone)
	if err != nil {
		return errors.New("receiver not found (invalid phone number)")
	}

	if sender.ID == receiver.ID {
		return errors.New("cannot transfer to yourself")
	}

	// 4. เริ่มต้น Database Transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)
	amountStr := fmt.Sprintf("%.2f", req.Amount)

	// A. หักเงินคนโอน
	err = qtx.DeductBalance(ctx, sqlc.DeductBalanceParams{
		WalletBalance: amountStr,
		ID:            sender.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to deduct sender balance: %v", err)
	}

	// B. เพิ่มเงินคนรับ
	err = qtx.AddBalance(ctx, sqlc.AddBalanceParams{
		WalletBalance: amountStr,
		ID:            receiver.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to add receiver balance: %v", err)
	}

	// C. บันทึกประวัติฝั่งคนโอน
	transferOutID := uuid.New()
	err = qtx.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:         transferOutID,
		UserID:     sender.ID,
		Amount:     "-" + amountStr,
		Type:       "transfer_out",
		Status:     "completed",
		ReferenceID: uuid.NullUUID{UUID: receiver.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to log sender transaction: %v", err)
	}

	// D. บันทึกประวัติฝั่งคนรับ
	transferInID := uuid.New()
	err = qtx.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:         transferInID,
		UserID:     receiver.ID,
		Amount:     "+" + amountStr,
		Type:       "transfer_in",
		Status:     "completed",
		ReferenceID: uuid.NullUUID{UUID: sender.ID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to log receiver transaction: %v", err)
	}

	// 5. ยืนยันการทำรายการทั้งหมด
	return tx.Commit()
}
