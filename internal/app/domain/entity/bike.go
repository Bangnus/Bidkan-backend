package entity

import (
	"time"

	"github.com/google/uuid"
)

type BikeData struct {
	BikeID        string     `json:"bike_id"`
	HardwareID    string     `json:"hardware_id"`
	TenantID      *uuid.UUID `json:"tenant_id,omitempty"`
	Lat           float64    `json:"lat"`
	Lon           float64    `json:"lon"`
	Battery       int        `json:"battery"`
	Status        string     `json:"status"`
	ImageURL      string     `json:"image_url"`
	CurrentRideID *string    `json:"current_ride_id"`
	LastHeartbeat time.Time  `json:"last_heartbeat"`
}
