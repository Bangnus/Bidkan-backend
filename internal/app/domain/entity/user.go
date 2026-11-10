package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       *uuid.UUID `json:"tenant_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	PhoneNumber    string     `json:"phone_number"`
	Email          string     `json:"email"`
	Username       string     `json:"username"`
	Password       string     `json:"-"`
	WalletBalance  string     `json:"wallet_balance"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	ImageURL       string     `json:"image_url"`
	FcmToken       string     `json:"fcm_token,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
