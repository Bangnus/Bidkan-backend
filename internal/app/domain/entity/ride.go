package entity

import (
	"time"

	"github.com/google/uuid"
)

type Ride struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	BikeID          string
	TenantID        uuid.UUID
	StartTime       time.Time
	EndTime         *time.Time
	StartLat        float64
	StartLon        float64
	EndLat          *float64
	EndLon          *float64
	DistanceKm      float64
	UnlockFee       string
	RatePerMinute   string
	TotalFare       string
	PlatformFee     string
	PartnerEarnings string
	Status          string // ongoing, completed, cancelled
	Type            string // ride, service
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
