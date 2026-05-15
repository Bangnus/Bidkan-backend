package topup

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/pkg/payment"
	"github.com/google/uuid"
)

type Request struct {
	UserID uuid.UUID `json:"-"`
	Amount float64   `json:"amount" validate:"required,gt=0"`
}

type Response struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	QRCodeURL     string    `json:"qr_code_url"`
}

type Service interface {
	RequestTopup(ctx context.Context, req Request) (*Response, error)
}

type service struct {
	txRepo    repository.TransactionRepository
	userRepo  repository.UserRepository
	paySvc    payment.PaySolutionsService
}

func NewService(txRepo repository.TransactionRepository, userRepo repository.UserRepository, paySvc payment.PaySolutionsService) Service {
	return &service{
		txRepo:   txRepo,
		userRepo: userRepo,
		paySvc:   paySvc,
	}
}

func (s *service) RequestTopup(ctx context.Context, req Request) (*Response, error) {
	// ตรวจสอบ User
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// สร้าง Transaction ในฐานข้อมูล (สถานะ pending)
	txID := uuid.New()
	amountStr := fmt.Sprintf("%.2f", req.Amount)

	tx := &entity.Transaction{
		ID:     txID,
		UserID: req.UserID,
		Amount: amountStr,
		Type:   "topup",
		Status: "pending",
	}

	err = s.txRepo.Create(ctx, tx)
	if err != nil {
		return nil, errors.New("failed to create transaction")
	}

	// ขอ QR Code จาก PaySolutions
	qrURL, err := s.paySvc.GeneratePromptPayQR(ctx, txID.String(), req.Amount)
	if err != nil {
		// ถ้าการขอ QR ล้มเหลว ให้อัปเดต Transaction เป็น failed
		_ = s.txRepo.UpdateStatus(ctx, txID, "failed", nil)
		return nil, errors.New("failed to generate promptpay qr code")
	}

	return &Response{
		TransactionID: txID,
		Amount:        req.Amount,
		QRCodeURL:     qrURL,
	}, nil
}
