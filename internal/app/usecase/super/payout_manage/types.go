package payout_manage

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type ServiceError struct {
	Status int
	Code   string
	Msg    string
	Payout *entity.TenantPayout
}

func (e *ServiceError) Error() string { return e.Msg }

type TriggerPayoutRequest struct {
	TenantSlug    string  `json:"tenant_slug"`
	GrossAmount   float64 `json:"gross_amount"`
	TotalRides    int     `json:"total_rides"`
	AutoCalculate bool    `json:"auto_calculate"`
	IncludeSimFee bool    `json:"include_sim_fee"`
}

type UnsettledTenantSummary struct {
	TenantID        uuid.UUID         `json:"tenant_id"`
	TenantSlug      string            `json:"tenant_slug"`
	TenantName      string            `json:"tenant_name"`
	GPTier          entity.GPTierInfo `json:"gp_tier"`
	HasStripe       bool              `json:"has_stripe"`
	StripeAccountID *string           `json:"stripe_account_id"`
	CampusBikes     int               `json:"campus_bikes"`
	MonthlySimFee   float64           `json:"monthly_sim_fee"`
	UnsettledRides  int               `json:"unsettled_rides"`
	UnsettledGross  float64           `json:"unsettled_gross"`
	UnsettledGPFee  float64           `json:"unsettled_gp_fee"`
	UnsettledNet    float64           `json:"unsettled_net"`
}

type AutoPayoutResult struct {
	Attempted int                    `json:"attempted"`
	Succeeded int                    `json:"succeeded"`
	Failed    int                    `json:"failed"`
	Skipped   int                    `json:"skipped"`
	Payouts   []*entity.TenantPayout `json:"payouts"`
	Errors    []string               `json:"errors,omitempty"`
}

type Service interface {
	ListOverview(ctx context.Context) ([]*entity.TenantPayout, []UnsettledTenantSummary, bool, bool, error)
	TriggerPayout(ctx context.Context, req TriggerPayoutRequest) (*entity.TenantPayout, error)
	RetryPayout(ctx context.Context, payoutID uuid.UUID) (*entity.TenantPayout, error)
	RunAutoPayouts(ctx context.Context) (*AutoPayoutResult, error)
}

