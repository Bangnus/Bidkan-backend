package entity

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	ContactEmail string    `json:"contact_email,omitempty"`
	ContactPhone string    `json:"contact_phone,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type HardwareOrder struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	TenantSlug  string     `json:"tenant_slug,omitempty"`
	TenantName  string     `json:"tenant_name,omitempty"`
	OrderType   string     `json:"order_type"`
	Quantity    int32      `json:"quantity"`
	UnitPrice   float64    `json:"unit_price"`
	TotalAmount float64    `json:"total_amount"`
	Note        string     `json:"note,omitempty"`
	Status      string     `json:"status"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type RideStatsSummary struct {
	TotalRides  int     `json:"total_rides"`
	GrossAmount float64 `json:"gross_amount"`
	GPFeeAmount float64 `json:"gp_fee_amount"`
	NetAmount   float64 `json:"net_amount"`
}
