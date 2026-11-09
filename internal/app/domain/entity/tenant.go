package entity

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID                        uuid.UUID  `json:"id"`
	OrganizationID            *uuid.UUID `json:"organization_id,omitempty"`
	Slug                      string     `json:"slug"`
	Name                      string     `json:"name"`
	StripeAccountID           *string    `json:"stripe_account_id"`
	StripeOnboardingCompleted bool       `json:"stripe_onboarding_completed"`
	PlatformFeePercent        float64    `json:"platform_fee_percent"`
	AutoTieredGP              bool       `json:"auto_tiered_gp"`
	UnlockFee                 float64    `json:"unlock_fee"`
	PerMinuteRate             float64    `json:"per_minute_rate"`
	ParkingPenaltyFee         float64    `json:"parking_penalty_fee"`
	SimFeePerBike             float64    `json:"sim_fee_per_bike"`
	ContactName               string     `json:"contact_name,omitempty"`
	ContactEmail              string     `json:"contact_email,omitempty"`
	ContactPhone              string     `json:"contact_phone,omitempty"`
	Notes                     string     `json:"notes,omitempty"`
	Status                    string     `json:"status"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
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

type BikeWithTenant struct {
	BikeID        string     `json:"bike_id"`
	HardwareID    string     `json:"hardware_id"`
	TenantID      *uuid.UUID `json:"tenant_id"`
	TenantSlug    string     `json:"tenant_slug"`
	TenantName    string     `json:"tenant_name"`
	Status        string     `json:"status"`
	Lat           float64    `json:"lat"`
	Lon           float64    `json:"lon"`
	Battery       int        `json:"battery"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CampusRideDetail struct {
	ID              uuid.UUID  `json:"id"`
	BikeID          string     `json:"bike_id"`
	UserID          uuid.UUID  `json:"user_id"`
	Username        string     `json:"username"`
	PhoneNumber     string     `json:"phone_number"`
	DistanceKm      float64    `json:"distance_km"`
	TotalFare       float64    `json:"total_fare"`
	PlatformFee     float64    `json:"platform_fee"`
	PartnerEarnings float64    `json:"partner_earnings"`
	Status          string     `json:"status"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
}

type MaintenanceReport struct {
	ID         uuid.UUID  `json:"id"`
	BikeID     string     `json:"bike_id"`
	ReportedBy uuid.UUID  `json:"reported_by"`
	IssueType  string     `json:"issue_type"`
	Status     string     `json:"status"`
	ResolvedBy *uuid.UUID `json:"resolved_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CampusStaff struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type TenantZone struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Boundary  string    `json:"boundary"`
	Radius    float64   `json:"radius"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
