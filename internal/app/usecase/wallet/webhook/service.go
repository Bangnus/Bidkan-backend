package webhook

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type PaySolutionsWebhookPayload struct {
	OrderNo string  `json:"order_no"`
	RefNo   string  `json:"ref_no"` // รหัสอ้างอิงของระบบ Payment
	Status  string  `json:"status"` // เช่น "CP" (Completed), "FA" (Failed)
	Amount  float64 `json:"amount"`
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
	txID, err := uuid.Parse(payload.OrderNo)
	if err != nil {
		return errors.New("invalid order_no format")
	}

	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil || tx == nil {
		return errors.New("transaction not found")
	}

	if tx.Status == "completed" {
		return nil
	}
	if tx.Status != "pending" {
		return errors.New("transaction is no longer pending")
	}

	expectedAmount, err := strconv.ParseFloat(tx.Amount, 64)
	if err != nil || expectedAmount <= 0 {
		return errors.New("invalid stored transaction amount")
	}

	if payload.Status == "CP" {
		if math.Abs(expectedAmount-payload.Amount) > 0.01 {
			return errors.New("webhook amount does not match transaction")
		}

		if err := s.txRepo.UpdateStatus(ctx, tx.ID, "completed", &payload.RefNo); err != nil {
			return errors.New("failed to update transaction status")
		}

		// Atomic DB increment using stored expected transaction amount
		if err := s.userRepo.AddBalance(ctx, tx.UserID, tx.Amount); err != nil {
			return errors.New("failed to credit user balance")
		}
		return nil
	}

	_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", &payload.RefNo)
	return errors.New("payment failed")
}
