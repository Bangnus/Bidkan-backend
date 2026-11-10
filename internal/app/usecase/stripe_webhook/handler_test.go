package stripe_webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type stubTenantRepo struct {
	domainRepo.TenantRepository
	tenant      *entity.Tenant
	updatedDone bool
}

func (s *stubTenantRepo) GetByStripeAccountID(_ context.Context, id string) (*entity.Tenant, error) {
	if s.tenant != nil && s.tenant.StripeAccountID != nil && *s.tenant.StripeAccountID == id {
		return s.tenant, nil
	}
	return nil, fmt.Errorf("not found")
}
func (s *stubTenantRepo) UpdateStripeStatus(_ context.Context, _ uuid.UUID, _ string, completed bool) error {
	s.updatedDone = completed
	return nil
}

type stubSuperRepo struct {
	domainRepo.SuperAdminRepository
	seenEvents map[string]bool
}

func (s *stubSuperRepo) TryRecordStripeEvent(_ context.Context, id, _ string) (bool, error) {
	if s.seenEvents[id] {
		return false, nil
	}
	s.seenEvents[id] = true
	return true, nil
}

type stubTxRepo struct {
	domainRepo.TransactionRepository
	tx        *entity.Transaction
	completed bool
}

func (s *stubTxRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Transaction, error) {
	if s.tx != nil && s.tx.ID == id {
		return s.tx, nil
	}
	return nil, fmt.Errorf("not found")
}
func (s *stubTxRepo) CompleteIfPending(_ context.Context, id uuid.UUID, ref *string) (bool, error) {
	if s.tx != nil && s.tx.ID == id && s.tx.Status == "pending" {
		s.tx.Status = "completed"
		s.tx.GatewayRef = ref
		s.completed = true
		return true, nil
	}
	return false, nil
}

type stubUserRepo struct {
	domainRepo.UserRepository
	creditedCount int
	creditedAmt   string
}

func (s *stubUserRepo) AddBalance(_ context.Context, _ uuid.UUID, amt string) error {
	s.creditedCount++
	s.creditedAmt = amt
	return nil
}

func signStripeBody(body []byte, secret string) string {
	ts := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.%s", ts, body)))
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestStripeWebhookCheckoutReplayAndConnect(t *testing.T) {
	txID := uuid.New()
	uID := uuid.New()
	acct := "acct_live_ku"
	tenantRepo := &stubTenantRepo{tenant: &entity.Tenant{ID: uuid.New(), Slug: "ku", StripeAccountID: &acct}}
	superRepo := &stubSuperRepo{seenEvents: map[string]bool{}}
	txRepo := &stubTxRepo{tx: &entity.Transaction{ID: txID, UserID: uID, Amount: "150.00", Status: "pending"}}
	userRepo := &stubUserRepo{}

	secret := "whsec_unit_secret"
	client := infraStripe.NewClient(&infraStripe.Config{SecretKey: "sk_test", WebhookSecret: secret})
	svc := NewService(tenantRepo, superRepo, txRepo, userRepo)
	h := NewHandler(svc, client)

	app := fiber.New()
	app.Post("/webhooks/stripe", h.Handle)

	// 1. Missing signature -> 401
	bodyCheckout := []byte(fmt.Sprintf(
		`{"id":"evt_pay_1","type":"checkout.session.completed","data":{"object":{"id":"cs_1","payment_status":"paid","amount_total":15000,"client_reference_id":"%s"}}}`,
		txID,
	))
	reqBad := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(bodyCheckout))
	if resp, _ := app.Test(reqBad); resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("missing sig got %d, want 401", resp.StatusCode)
	}

	// 2. Valid Checkout Paid + Replay twice -> credited ONLY ONCE
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(bodyCheckout))
		req.Header.Set("Stripe-Signature", signStripeBody(bodyCheckout, secret))
		resp, _ := app.Test(req)
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("valid webhook try %d got %d", i, resp.StatusCode)
		}
	}
	if userRepo.creditedCount != 1 || userRepo.creditedAmt != "150.00" {
		t.Fatalf("expected 1 credit of 150.00, got count=%d amt=%s", userRepo.creditedCount, userRepo.creditedAmt)
	}

	// 3. Connect account.updated -> marks tenant onboarded
	bodyAcct := []byte(`{"id":"evt_acct_1","type":"account.updated","data":{"object":{"id":"acct_live_ku","payouts_enabled":true}}}`)
	reqAcct := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(bodyAcct))
	reqAcct.Header.Set("Stripe-Signature", signStripeBody(bodyAcct, secret))
	respAcct, _ := app.Test(reqAcct)
	if respAcct.StatusCode != fiber.StatusOK || !tenantRepo.updatedDone {
		t.Fatalf("account.updated failed status=%d updatedDone=%v", respAcct.StatusCode, tenantRepo.updatedDone)
	}
}

