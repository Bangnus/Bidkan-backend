package entity

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID                        uuid.UUID `json:"id"`
	Slug                      string    `json:"slug"`
	Name                      string    `json:"name"`
	StripeAccountID           *string   `json:"stripe_account_id"`
	StripeOnboardingCompleted bool      `json:"stripe_onboarding_completed"`
	PlatformFeePercent        float64   `json:"platform_fee_percent"`
	UnlockFee                 float64   `json:"unlock_fee"`
	PerMinuteRate             float64   `json:"per_minute_rate"`
	Status                    string    `json:"status"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type TenantPayout struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	StripeTransferID  *string    `json:"stripe_transfer_id"`
	TotalRides        int        `json:"total_rides"`
	GrossAmount       float64    `json:"gross_amount"`
	PlatformFeeAmount float64    `json:"platform_fee_amount"`
	NetPayoutAmount   float64    `json:"net_payout_amount"`
	Status            string     `json:"status"`
	ErrorMessage      *string    `json:"error_message"`
	PeriodStart       time.Time  `json:"period_start"`
	PeriodEnd         time.Time  `json:"period_end"`
	ProcessedAt       *time.Time `json:"processed_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

