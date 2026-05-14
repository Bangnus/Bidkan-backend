package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	PhoneNumber string
	FullName    string
	Password    string
	WalletBalance string
	Role        string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}