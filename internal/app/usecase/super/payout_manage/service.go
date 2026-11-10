package payout_manage

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	infraStripe "github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/stripe"
	"github.com/google/uuid"
)

type service struct {
	superRepo  domainRepo.SuperAdminRepository
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
	stripe     infraStripe.Client
}

func NewService(
	superRepo domainRepo.SuperAdminRepository,
	tenantRepo domainRepo.TenantRepository,
	campusRepo domainRepo.CampusOpsRepository,
	stripeClient infraStripe.Client,
) Service {
	if stripeClient == nil {
		stripeClient = infraStripe.NewClient(nil)
	}
	return &service{superRepo: superRepo, tenantRepo: tenantRepo, campusRepo: campusRepo, stripe: stripeClient}
}

func (s *service) ListOverview(ctx context.Context) ([]*entity.TenantPayout, []UnsettledTenantSummary, bool, bool, error) {
	payouts, err := s.superRepo.ListAllPayouts(ctx)
	if err != nil {
		return nil, nil, false, false, err
	}
	tenants, _ := s.tenantRepo.List(ctx)
	unsettled := make([]UnsettledTenantSummary, 0, len(tenants))
	for _, t := range tenants {
		st, _ := s.campusRepo.GetUnsettledStats(ctx, t.ID)
		if st == nil {
			st = &entity.RideStatsSummary{}
		}
		bikes, _ := s.tenantRepo.ListBikesByTenant(ctx, t.ID)
		fleetCnt, _ := s.tenantRepo.CountFleetForTenant(ctx, t.ID)
		unsettled = append(unsettled, UnsettledTenantSummary{
			TenantID:        t.ID,
			TenantSlug:      t.Slug,
			TenantName:      t.Name,
			GPTier:          entity.CalculateTieredGP(fleetCnt),
			HasStripe:       t.StripeOnboardingCompleted && t.StripeAccountID != nil && *t.StripeAccountID != "",
			StripeAccountID: t.StripeAccountID,
			CampusBikes:     len(bikes),
			MonthlySimFee:   float64(len(bikes)) * t.SimFeePerBike,
			UnsettledRides:  st.TotalRides,
			UnsettledGross:  st.GrossAmount,
			UnsettledGPFee:  st.GPFeeAmount,
			UnsettledNet:    st.NetAmount,
		})
	}
	return payouts, unsettled, s.stripe.IsConfigured(), s.stripe.Config().AutoPayoutEnabled, nil
}

func (s *service) TriggerPayout(ctx context.Context, req TriggerPayoutRequest) (*entity.TenantPayout, error) {
	if !s.stripe.IsConfigured() {
		return nil, &ServiceError{Status: http.StatusServiceUnavailable, Code: "STRIPE_NOT_CONFIGURED", Msg: "Stripe ยังไม่ได้ตั้งค่า (STRIPE_SECRET_KEY)"}
	}
	tenant, err := s.tenantRepo.GetBySlug(ctx, req.TenantSlug)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Msg: "tenant not found"}
	}
	if !tenant.StripeOnboardingCompleted || tenant.StripeAccountID == nil || *tenant.StripeAccountID == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "STRIPE_NOT_ONBOARDED", Msg: "วิทยาเขตนี้ยังไม่ได้เชื่อมต่อและยืนยันบัญชี Stripe Connect"}
	}
	if inFlight, err := s.superRepo.HasInFlightPayout(ctx, tenant.ID); err == nil && inFlight {
		return nil, &ServiceError{Status: http.StatusConflict, Code: "PAYOUT_IN_FLIGHT", Msg: "มีรายการโอนเงินที่กำลังดำเนินการอยู่สำหรับวิทยาเขตนี้"}
	}

	now := time.Now()
	periodStart, periodEnd := now.AddDate(0, 0, -7), now
	gross, rides := req.GrossAmount, req.TotalRides
	gpFee := math.Round((gross*(tenant.PlatformFeePercent/100.0))*100) / 100
	if req.AutoCalculate || gross <= 0 {
		st, err := s.campusRepo.GetUnsettledPeriod(ctx, tenant.ID)
		if err != nil || st.GrossAmount <= 0 || st.TotalRides == 0 {
			return nil, &ServiceError{Status: http.StatusBadRequest, Msg: "ยังไม่มีรายได้ค้างโอนจากทริปที่จบแล้วในรอบบัญชีนี้"}
		}
		gross, rides, gpFee = st.GrossAmount, st.TotalRides, st.GPFeeAmount
		if st.PeriodStart != nil && st.PeriodEnd != nil {
			periodStart, periodEnd = *st.PeriodStart, *st.PeriodEnd
		}
	}
	if req.IncludeSimFee && tenant.SimFeePerBike > 0 {
		bikes, _ := s.tenantRepo.ListBikesByTenant(ctx, tenant.ID)
		gpFee += float64(len(bikes)) * tenant.SimFeePerBike
	}
	netAmount := math.Round((gross-gpFee)*100) / 100
	if netAmount < 1.00 {
		return nil, &ServiceError{Status: http.StatusBadRequest, Msg: "ยอดสุทธิหลังหักค่าบริการน้อยกว่า 1.00 บาท ไม่สามารถโอนผ่าน Stripe ได้"}
	}

	payout := &entity.TenantPayout{
		ID:                uuid.New(),
		TenantID:          tenant.ID,
		TotalRides:        rides,
		GrossAmount:       gross,
		PlatformFeeAmount: gpFee,
		NetPayoutAmount:   netAmount,
		Status:            "processing",
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		CreatedAt:         now,
	}
	if err := s.superRepo.CreatePayout(ctx, payout); err != nil {
		return nil, err
	}
	return s.dispatchTransfer(ctx, payout, tenant, "payout-"+payout.ID.String())
}

func (s *service) dispatchTransfer(ctx context.Context, payout *entity.TenantPayout, tenant *entity.Tenant, idempotencyKey string) (*entity.TenantPayout, error) {
	res, err := s.stripe.CreateTransfer(ctx, *tenant.StripeAccountID, payout.NetPayoutAmount, idempotencyKey, map[string]string{
		"payout_id":   payout.ID.String(),
		"tenant_id":   tenant.ID.String(),
		"tenant_slug": tenant.Slug,
	})
	now := time.Now()
	if err != nil {
		msg := err.Error()
		_ = s.superRepo.UpdatePayoutStatus(ctx, payout.ID, "failed", nil, &msg)
		payout.Status = "failed"
		payout.ErrorMessage = &msg
		payout.ProcessedAt = &now
		return nil, &ServiceError{Status: http.StatusBadGateway, Code: "STRIPE_TRANSFER_FAILED", Msg: msg, Payout: payout}
	}
	_ = s.superRepo.UpdatePayoutStatus(ctx, payout.ID, "completed", &res.ID, nil)
	payout.Status = "completed"
	payout.StripeTransferID = &res.ID
	payout.ErrorMessage = nil
	payout.ProcessedAt = &now
	return payout, nil
}
