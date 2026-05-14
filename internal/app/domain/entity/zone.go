package entity

import (
	"time"
	"github.com/google/uuid"
)

type Zone struct {
	ID        uuid.UUID
	Name      string
	Type      string
	Boundary  string
	Radius    float64
	CreatedAt time.Time
	UpdatedAt time.Time
}