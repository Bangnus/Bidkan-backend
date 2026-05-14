package entity

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Amount      string
	Type        string
	ReferenceID uuid.NullUUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}