package stripe_topup

import (
	"context"
	"errors"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/google/uuid"
)

type stubTxRepo struct {
	domainRepo.TransactionRepository
	items map[uuid.UUID]*entity.Transaction
}

func (s *stubTxRepo) Create(_ context.Context, tx *entity.Transaction) error {
	cp := *tx
	s.items[tx.ID] = &cp
	return nil
}
func (s *stubTxRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Transaction, error) {
	if tx, ok := s.items[id]; ok {
		cp := *tx
		return &cp, nil
	}
	return nil, errors.New("not found")
}
func (s *stubTxRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string, ref *string) error {
	if tx, ok := s.items[id]; ok {
		tx.Status = status
		tx.GatewayRef = ref
	}
	return nil
}
func (s *stubTxRepo) CompleteIfPending(_ context.Context, id uuid.UUID, ref *string) (bool, error) {
	if tx, ok := s.items[id]; ok && tx.Status == "pending" {
		tx.Status = "completed"
		tx.GatewayRef = ref
		return true, nil
	}
	return false, nil
}

type stubUserRepo struct {
	domainRepo.UserRepository
	user         *entity.User
	creditCount  int
	lastCredited string
}

func (s *stubUserRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	if s.user != nil && s.user.ID == id {
		return s.user, nil
	}
	return nil, errors.New("not found")
}
func (s *stubUserRepo) AddBalance(_ context.Context, _ uuid.UUID, amt string) error {
	s.creditCount++
	s.lastCredited = amt
	return nil
}

type stubStripeClient struct {
	infraStripe.Client
	configured bool
	session    *infraStripe.CheckoutSession
}

func (s *stubStripeClient) IsConfigured() bool { return s.configured }
func (s *stubStripeClient) Config() *infraStripe.Config {
	return &infraStripe.Config{ClientURL: "http://localhost:3000"}
}
func (s *stubStripeClient) CreateCheckoutSession(_ context.Context, p infraStripe.CheckoutSessionParams) (*infraStripe.CheckoutSession, error) {
	s.session = &infraStripe.CheckoutSession{
		ID:                "cs_unit_1",
		URL:               "https://checkout.stripe.com/pay/cs_unit_1",
		PaymentStatus:     "paid",
		AmountTotal:       int64(p.AmountTHB * 100),
		ClientReferenceID: p.TransactionID,
	}
	return s.session, nil
}
func (s *stubStripeClient) GetCheckoutSession(context.Context, string) (*infraStripe.CheckoutSession, error) {
	return s.session, nil
}

func TestStripeTopupCreateVerifyAndIdempotency(t *testing.T) {
	uID := uuid.New()
	txRepo := &stubTxRepo{items: map[uuid.UUID]*entity.Transaction{}}
	userRepo := &stubUserRepo{user: &entity.User{ID: uID, Email: "rider@ku.th"}}
	st := &stubStripeClient{configured: false}
	svc := NewService(txRepo, userRepo, st)

	if _, err := svc.CreateCheckout(context.Background(), CheckoutRequest{UserID: uID, Amount: 100}); err == nil {
		t.Fatalf("expected error when stripe unconfigured")
	}

	st.configured = true
	res, err := svc.CreateCheckout(context.Background(), CheckoutRequest{UserID: uID, Amount: 200})
	if err != nil || res.SessionID != "cs_unit_1" {
		t.Fatalf("CreateCheckout got (%+v, %v)", res, err)
	}

	// IDOR: different user cannot verify another rider's session
	if _, err := svc.VerifySession(context.Background(), uuid.New(), "cs_unit_1"); err == nil {
		t.Fatalf("expected forbidden when another user verifies session")
	}

	// First verify credits wallet; second verify does not double-credit
	v1, err := svc.VerifySession(context.Background(), uID, "cs_unit_1")
	if err != nil || !v1.Credited || v1.Status != "completed" {
		t.Fatalf("VerifySession #1 got (%+v, %v)", v1, err)
	}
	v2, err := svc.VerifySession(context.Background(), uID, "cs_unit_1")
	if err != nil || v2.Credited || userRepo.creditCount != 1 {
		t.Fatalf("VerifySession #2 got (%+v, err=%v, credits=%d)", v2, err, userRepo.creditCount)
	}
}

