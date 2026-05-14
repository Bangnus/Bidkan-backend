package entity

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID         uuid.UUID
	BikeID     string
	ReportedBy uuid.UUID
	IssueType  string
	Status     string
	ResolvedBy uuid.NullUUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
