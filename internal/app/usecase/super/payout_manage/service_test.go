package payout_manage

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/google/uuid"
)

type stubSuperRepo struct {
	domainRepo.SuperAdminRepository
	payouts  map[uuid.UUID]*entity.TenantPayout
	inFlight bool
}

func (s *stubSuperRepo) CreatePayout(_ context.Context, p *entity.TenantPayout) error {
	cp := *p
	s.payouts[p.ID] = &cp
	return nil
}
func (s *stubSuperRepo) GetPayoutByID(_ context.Context, id uuid.UUID) (*entity.TenantPayout, error) {
	if p, ok := s.payouts[id]; ok {
		cp := *p
		return &cp, nil
	}
	return nil, errors.New("not found")
}
func (s *stubSuperRepo) HasInFlightPayout(context.Context, uuid.UUID) (bool, error) {
	return s.inFlight, nil
}
func (s *stubSuperRepo) UpdatePayoutStatus(_ context.Context, id uuid.UUID, status string, trID, errMsg *string) error {
	if p, ok := s.payouts[id]; ok {
		p.Status = status
		p.StripeTransferID = trID
		p.ErrorMessage = errMsg
	}
	return nil
}

type stubTenantRepo struct {
	domainRepo.TenantRepository
	tenants []*entity.Tenant
}

func (s *stubTenantRepo) List(context.Context) ([]*entity.Tenant, error) { return s.tenants, nil }
func (s *stubTenantRepo) GetBySlug(_ context.Context, slug string) (*entity.Tenant, error) {
	for _, t := range s.tenants {
		if t.Slug == slug {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}
func (s *stubTenantRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Tenant, error) {
	for _, t := range s.tenants {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}
func (s *stubTenantRepo) ListBikesByTenant(context.Context, uuid.UUID) ([]*entity.BikeData, error) {
	return []*entity.BikeData{{BikeID: "B1"}}, nil
}

type stubCampusOps struct {
	domainRepo.CampusOpsRepository
	period *entity.RideStatsSummary
}

func (s *stubCampusOps) GetUnsettledPeriod(context.Context, uuid.UUID) (*entity.RideStatsSummary, error) {
	return s.period, nil
}

type stubStripeClient struct {
	infraStripe.Client
	configured bool
	transferID string
	failNext   error
}

func (s *stubStripeClient) IsConfigured() bool        { return s.configured }
func (s *stubStripeClient) Config() *infraStripe.Config { return &infraStripe.Config{} }
func (s *stubStripeClient) CreateTransfer(_ context.Context, dest string, amt float64, _ string, _ map[string]string) (*infraStripe.TransferResult, error) {
	if s.failNext != nil {
		err := s.failNext
		s.failNext = nil
		return nil, err
	}
	return &infraStripe.TransferResult{ID: s.transferID, Amount: int64(amt * 100), Destination: dest}, nil
}

func TestTriggerAndRetryPayout(t *testing.T) {
	acct := "acct_123"
	now := time.Now()
	tID := uuid.New()
	superRepo := &stubSuperRepo{payouts: map[uuid.UUID]*entity.TenantPayout{}}
	tenantRepo := &stubTenantRepo{tenants: []*entity.Tenant{{
		ID: tID, Slug: "ku", Name: "KU", Status: "active", StripeAccountID: &acct, StripeOnboardingCompleted: true,
	}}}
	campusRepo := &stubCampusOps{period: &entity.RideStatsSummary{
		TotalRides: 5, GrossAmount: 500, GPFeeAmount: 75, NetAmount: 425, PeriodStart: &now, PeriodEnd: &now,
	}}
	st := &stubStripeClient{configured: false, transferID: "tr_real_1"}
	svc := NewService(superRepo, tenantRepo, campusRepo, st)

	if _, err := svc.TriggerPayout(context.Background(), TriggerPayoutRequest{TenantSlug: "ku", AutoCalculate: true}); err == nil {
		t.Fatalf("expected error when stripe unconfigured")
	}

	st.configured = true
	st.failNext = &infraStripe.APIError{StatusCode: http.StatusBadGateway, Message: "insufficient funds"}
	if _, err := svc.TriggerPayout(context.Background(), TriggerPayoutRequest{TenantSlug: "ku", AutoCalculate: true}); err == nil {
		t.Fatalf("expected failed transfer error")
	}
	var failedID uuid.UUID
	for id, p := range superRepo.payouts {
		if p.Status == "failed" {
			failedID = id
		}
	}
	if failedID == uuid.Nil {
		t.Fatalf("expected payout saved as failed")
	}

	retried, err := svc.RetryPayout(context.Background(), failedID)
	if err != nil || retried.Status != "completed" || *retried.StripeTransferID != "tr_real_1" {
		t.Fatalf("RetryPayout got (%+v, %v)", retried, err)
	}
}

