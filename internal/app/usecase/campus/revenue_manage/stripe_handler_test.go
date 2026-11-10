package revenue_manage

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type stubTenantRepo struct {
	domainRepo.TenantRepository
	tenant      *entity.Tenant
	updatedAcct string
	updatedDone bool
}

func (s *stubTenantRepo) GetBySlug(_ context.Context, slug string) (*entity.Tenant, error) {
	if s.tenant != nil && s.tenant.Slug == slug {
		return s.tenant, nil
	}
	return nil, errors.New("not found")
}
func (s *stubTenantRepo) UpdateStripeStatus(_ context.Context, _ uuid.UUID, acct string, done bool) error {
	s.updatedAcct = acct
	s.updatedDone = done
	s.tenant.StripeAccountID = &acct
	s.tenant.StripeOnboardingCompleted = done
	return nil
}

type stubStripeClient struct {
	infraStripe.Client
	configured bool
	acctStatus *infraStripe.AccountStatus
}

func (s *stubStripeClient) IsConfigured() bool { return s.configured }
func (s *stubStripeClient) Config() *infraStripe.Config {
	return &infraStripe.Config{ConnectAccountType: "express", BackofficeURL: "http://localhost:3001"}
}
func (s *stubStripeClient) CreateConnectedAccount(context.Context, string, string, string) (string, error) {
	return "acct_live_new", nil
}
func (s *stubStripeClient) GetAccount(_ context.Context, id string) (*infraStripe.AccountStatus, error) {
	if id == "acct_missing" {
		return nil, &infraStripe.APIError{StatusCode: 404, Code: "resource_missing", Message: "missing"}
	}
	return s.acctStatus, nil
}
func (s *stubStripeClient) CreateAccountLink(_ context.Context, acctID, _, _ string) (string, error) {
	return "https://connect.stripe.com/setup/" + acctID, nil
}
func (s *stubStripeClient) CreateExpressLoginLink(context.Context, string) (string, error) {
	return "https://connect.stripe.com/express/login", nil
}

func TestStripeConnectOnboardAndStatusSync(t *testing.T) {
	tenant := &entity.Tenant{ID: uuid.New(), Slug: "ku", Name: "KU", ContactEmail: "ops@ku.th"}
	repo := &stubTenantRepo{tenant: tenant}
	st := &stubStripeClient{configured: false}
	h := NewHandler(repo, nil, st)

	app := fiber.New()
	app.Get("/campus/:slug/stripe/status", h.HandleGetStripeStatus)
	app.Post("/campus/:slug/stripe/onboard", h.HandleStripeOnboard)
	app.Post("/campus/:slug/stripe/login-link", h.HandleStripeLoginLink)

	// Unconfigured returns 503
	if resp, _ := app.Test(httptest.NewRequest("POST", "/campus/ku/stripe/onboard", nil)); resp.StatusCode != 503 {
		t.Fatalf("expected 503 when unconfigured, got %d", resp.StatusCode)
	}

	// Configured creates account + AccountLink URL
	st.configured = true
	resp, _ := app.Test(httptest.NewRequest("POST", "/campus/ku/stripe/onboard", nil))
	if resp.StatusCode != 200 || repo.updatedAcct != "acct_live_new" || repo.updatedDone {
		t.Fatalf("onboard failed status=%d acct=%s done=%v", resp.StatusCode, repo.updatedAcct, repo.updatedDone)
	}
	var payload map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if payload["onboarding_url"] == "" {
		t.Fatalf("expected onboarding_url in response")
	}

	// Status sync marks onboarding_completed when payouts_enabled=true
	st.acctStatus = &infraStripe.AccountStatus{ID: "acct_live_new", PayoutsEnabled: true, DetailsSubmitted: true}
	respSt, _ := app.Test(httptest.NewRequest("GET", "/campus/ku/stripe/status", nil))
	if respSt.StatusCode != 200 || !repo.updatedDone {
		t.Fatalf("status sync failed status=%d done=%v", respSt.StatusCode, repo.updatedDone)
	}
}

