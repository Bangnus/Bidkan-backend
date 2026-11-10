package stripe_topup

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/google/uuid"
)

type service struct {
	txRepo   domainRepo.TransactionRepository
	userRepo domainRepo.UserRepository
	stripe   infraStripe.Client
}

func NewService(
	txRepo domainRepo.TransactionRepository,
	userRepo domainRepo.UserRepository,
	stripeClient infraStripe.Client,
) Service {
	if stripeClient == nil {
		stripeClient = infraStripe.NewClient(nil)
	}
	return &service{txRepo: txRepo, userRepo: userRepo, stripe: stripeClient}
}

func (s *service) GetConfig() ConfigResponse {
	paySolReady := strings.TrimSpace(os.Getenv("PAYSOLUTIONS_MERCHANT_ID")) != "" &&
		strings.TrimSpace(os.Getenv("PAYSOLUTIONS_API_KEY")) != ""
	return ConfigResponse{
		StripeEnabled:       s.stripe.IsConfigured(),
		PaySolutionsEnabled: paySolReady,
		Currency:            "thb",
		MinAmount:           MinTopupTHB,
		MaxAmount:           MaxTopupTHB,
	}
}

func (s *service) CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error) {
	if !s.stripe.IsConfigured() {
		return nil, &ServiceError{Status: http.StatusServiceUnavailable, Code: "STRIPE_NOT_CONFIGURED", Msg: "ระบบชำระเงินผ่าน Stripe ยังไม่ได้ตั้งค่า (STRIPE_SECRET_KEY)"}
	}
	if req.Amount < MinTopupTHB || req.Amount > MaxTopupTHB {
		return nil, &ServiceError{Status: http.StatusBadRequest, Msg: fmt.Sprintf("จำนวนเงินต้องอยู่ระหว่าง %.0f ถึง %.0f บาท", MinTopupTHB, MaxTopupTHB)}
	}
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil || user == nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Msg: "user not found"}
	}

	txID := uuid.New()
	tx := &entity.Transaction{
		ID:     txID,
		UserID: req.UserID,
		Amount: fmt.Sprintf("%.2f", req.Amount),
		Type:   "topup",
		Status: "pending",
	}
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Msg: "failed to create transaction"}
	}

	clientURL := s.stripe.Config().ClientURL
	session, err := s.stripe.CreateCheckoutSession(ctx, infraStripe.CheckoutSessionParams{
		TransactionID: txID.String(),
		UserID:        req.UserID.String(),
		AmountTHB:     req.Amount,
		CustomerEmail: user.Email,
		Method:        strings.ToLower(strings.TrimSpace(req.Method)),
		SuccessURL:    clientURL + "/wallet?topup=success&session_id={CHECKOUT_SESSION_ID}",
		CancelURL:     clientURL + "/wallet?topup=cancelled",
	})
	if err != nil {
		_ = s.txRepo.UpdateStatus(ctx, txID, "failed", nil)
		return nil, &ServiceError{Status: http.StatusBadGateway, Code: "STRIPE_CHECKOUT_FAILED", Msg: err.Error()}
	}
	_ = s.txRepo.UpdateStatus(ctx, txID, "pending", &session.ID)

	return &CheckoutResponse{
		TransactionID: txID,
		SessionID:     session.ID,
		CheckoutURL:   session.URL,
		Amount:        req.Amount,
	}, nil
}

func (s *service) VerifySession(ctx context.Context, userID uuid.UUID, sessionID string) (*VerifyResponse, error) {
	if !s.stripe.IsConfigured() {
		return nil, &ServiceError{Status: http.StatusServiceUnavailable, Code: "STRIPE_NOT_CONFIGURED", Msg: "ระบบชำระเงินผ่าน Stripe ยังไม่ได้ตั้งค่า"}
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Msg: "session_id is required"}
	}
	session, err := s.stripe.GetCheckoutSession(ctx, sessionID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusBadGateway, Msg: err.Error()}
	}

	tx, err := s.resolveTransaction(ctx, *session)
	if err != nil || tx == nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Msg: "transaction not found"}
	}
	if tx.UserID != userID {
		return nil, &ServiceError{Status: http.StatusForbidden, Msg: "คุณไม่มีสิทธิ์เข้าถึงรายการเติมเงินนี้"}
	}

	amt, _ := strconv.ParseFloat(tx.Amount, 64)
	if session.PaymentStatus != "paid" {
		return &VerifyResponse{TransactionID: tx.ID, SessionID: session.ID, Status: tx.Status, Amount: amt, Credited: false}, nil
	}

	expectedSatang := int64(math.Round(amt * 100))
	if session.AmountTotal > 0 && expectedSatang-session.AmountTotal > 1 {
		return nil, &ServiceError{Status: http.StatusBadRequest, Msg: "จำนวนเงินจาก Stripe ไม่ตรงกับรายการเติมเงิน"}
	}

	completedNow, err := s.txRepo.CompleteIfPending(ctx, tx.ID, &session.ID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Msg: "failed to finalize transaction"}
	}
	if completedNow {
		if err := s.userRepo.AddBalance(ctx, tx.UserID, tx.Amount); err != nil {
			return nil, &ServiceError{Status: http.StatusInternalServerError, Msg: "failed to credit wallet"}
		}
	}
	return &VerifyResponse{
		TransactionID: tx.ID,
		SessionID:     session.ID,
		Status:        "completed",
		Amount:        amt,
		Credited:      completedNow,
	}, nil
}

func (s *service) resolveTransaction(ctx context.Context, session infraStripe.CheckoutSession) (*entity.Transaction, error) {
	txIDStr := session.Metadata["transaction_id"]
	if txIDStr == "" {
		txIDStr = session.ClientReferenceID
	}
	if txIDStr != "" {
		if id, err := uuid.Parse(txIDStr); err == nil {
			return s.txRepo.GetByID(ctx, id)
		}
	}
	return s.txRepo.GetByGatewayRef(ctx, session.ID)
}
