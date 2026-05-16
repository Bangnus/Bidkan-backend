package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	PhoneNumber   string
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