package transfer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
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
	mqttPub   mqtt.Publisher
	notiPub   service.NotificationProvider
	queries   *sqlc.Queries // สำหรับรันใน Transaction
}

func NewService(db *sql.DB, userRepo repository.UserRepository, txRepo repository.TransactionRepository, mqttPub mqtt.Publisher, notiPub service.NotificationProvider) Service {
	return &service{
		db:       db,
		userRepo: userRepo,
		txRepo:   txRepo,
		mqttPub:  mqttPub,
		notiPub:  notiPub,
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
	if err := tx.Commit(); err != nil {
		return err
	}

	// 6. ส่งแจ้งเตือน Real-time (ถ้าต่อ MQTT สำเร็จ)
	if s.mqttPub != nil {
		// แจ้งฝั่งคนโอน
		s.mqttPub.Publish(fmt.Sprintf("bidkan/users/%s/notifications", sender.ID), map[string]interface{}{
			"type":    "wallet_transfer_out",
			"amount":  req.Amount,
			"to":      receiver.Username,
			"balance": "updated", // หรือจะคำนวณ balance ใหม่ส่งไปเลยก็ได้
			"time":    time.Now().Format(time.RFC3339),
		})

		// แจ้งฝั่งคนรับ
		s.mqttPub.Publish(fmt.Sprintf("bidkan/users/%s/notifications", receiver.ID), map[string]interface{}{
			"type":    "wallet_transfer_in",
			"amount":  req.Amount,
			"from":    sender.Username,
			"balance": "updated",
			"time":    time.Now().Format(time.RFC3339),
		})
	}

	// 7. ส่ง Push Notification (FCM) หาคนรับ
	if s.notiPub != nil && receiver.FcmToken != "" {
		_ = s.notiPub.SendToToken(ctx, receiver.FcmToken, "ได้รับเงินโอน", fmt.Sprintf("คุณได้รับเงินจำนวน %s บาท จาก %s", req.Amount, sender.Username), map[string]string{
			"type": "wallet_transfer_in",
		})
	}

	return nil
}
