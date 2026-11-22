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
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	EmailVerified   bool       `json:"email_verified"`
	ImageURL        string     `json:"image_url"`
	FcmToken        string     `json:"fcm_token,omitempty"`
	IsStudent       bool       `json:"is_student"`
	StudentID       string     `json:"student_id,omitempty"`
	StudentCardURL  string     `json:"student_card_url,omitempty"`
	StudentVerified bool       `json:"student_verified"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
