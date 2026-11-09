package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	TenantID       *uuid.UUID
	OrganizationID *uuid.UUID
	PhoneNumber    string
	Email         string
	Username      string
	Password      string
	WalletBalance string
	Role          string
	Status        string
	ImageURL      string
	FcmToken      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}