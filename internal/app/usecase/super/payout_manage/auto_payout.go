package payout_manage

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

func (s *service) RetryPayout(ctx context.Context, payoutID uuid.UUID) (*entity.TenantPayout, error) {
	if !s.stripe.IsConfigured() {
		return nil, &ServiceError{Status: http.StatusServiceUnavailable, Code: "STRIPE_NOT_CONFIGURED", Msg: "Stripe ยังไม่ได้ตั้งค่า (STRIPE_SECRET_KEY)"}
	}
	payout, err := s.superRepo.GetPayoutByID(ctx, payoutID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Msg: "payout not found"}
	}
	if payout.Status != "failed" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Msg: "สามารถโอนซ้ำได้เฉพาะรายการที่สถานะ failed เท่านั้น"}
	}
	if inFlight, err := s.superRepo.HasInFlightPayout(ctx, payout.TenantID); err == nil && inFlight {
		return nil, &ServiceError{Status: http.StatusConflict, Code: "PAYOUT_IN_FLIGHT", Msg: "มีรายการโอนเงินที่กำลังดำเนินการอยู่สำหรับวิทยาเขตนี้"}
	}
	tenant, err := s.tenantRepo.GetByID(ctx, payout.TenantID)
	if err != nil || !tenant.StripeOnboardingCompleted || tenant.StripeAccountID == nil || *tenant.StripeAccountID == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Msg: "วิทยาเขตยังไม่ได้เชื่อมต่อบัญชี Stripe Connect ที่พร้อมรับโอน"}
	}
	if err := s.superRepo.UpdatePayoutStatus(ctx, payout.ID, "processing", nil, nil); err != nil {
		return nil, err
	}
	return s.dispatchTransfer(ctx, payout, tenant, fmt.Sprintf("payout-retry-%s-%s", payout.ID.String(), uuid.NewString()[:8]))
}

func (s *service) RunAutoPayouts(ctx context.Context) (*AutoPayoutResult, error) {
	if !s.stripe.IsConfigured() {
		return nil, &ServiceError{Status: http.StatusServiceUnavailable, Code: "STRIPE_NOT_CONFIGURED", Msg: "Stripe ยังไม่ได้ตั้งค่า (STRIPE_SECRET_KEY)"}
	}
	tenants, err := s.tenantRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	res := &AutoPayoutResult{Payouts: []*entity.TenantPayout{}}
	for _, t := range tenants {
		if t.Status != "active" || !t.StripeOnboardingCompleted || t.StripeAccountID == nil || *t.StripeAccountID == "" {
			res.Skipped++
			continue
		}
		if inFlight, _ := s.superRepo.HasInFlightPayout(ctx, t.ID); inFlight {
			res.Skipped++
			continue
		}
		st, err := s.campusRepo.GetUnsettledPeriod(ctx, t.ID)
		if err != nil || st.TotalRides == 0 || st.NetAmount < 1.00 {
			res.Skipped++
			continue
		}
		res.Attempted++
		payout, err := s.TriggerPayout(ctx, TriggerPayoutRequest{TenantSlug: t.Slug, AutoCalculate: true, IncludeSimFee: true})
		if err != nil {
			var svcErr *ServiceError
			if errors.As(err, &svcErr) && svcErr.Payout != nil {
				res.Payouts = append(res.Payouts, svcErr.Payout)
			}
			res.Failed++
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %s", t.Slug, err.Error()))
			continue
		}
		res.Succeeded++
		res.Payouts = append(res.Payouts, payout)
	}
	return res, nil
}

