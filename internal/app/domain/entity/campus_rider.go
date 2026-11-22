package entity

import (
	"time"

	"github.com/google/uuid"
)

type CampusRider struct {
	UserID          uuid.UUID `json:"user_id"`
	Username        string    `json:"username"`
	PhoneNumber     string    `json:"phone_number"`
	Email           string    `json:"email"`
	Status          string    `json:"status"` // active, suspended, pending
	WalletBalance   float64   `json:"wallet_balance"`
	IsStudent       bool      `json:"is_student"`
	StudentID       string    `json:"student_id"`
	StudentCardURL  string    `json:"student_card_url"`
	StudentVerified bool      `json:"student_verified"`
	TotalRides      int       `json:"total_rides"`
	DisputeCount    int       `json:"dispute_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type DisputedRide struct {
	RideID         uuid.UUID  `json:"ride_id"`
	UserID         uuid.UUID  `json:"user_id"`
	Username       string     `json:"username"`
	PhoneNumber    string     `json:"phone_number"`
	BikeID         string     `json:"bike_id"`
	TotalFare      float64    `json:"total_fare"`
	ParkingPenalty float64    `json:"parking_penalty"`
	PenaltyStatus  string     `json:"penalty_status"` // none, applied, disputed, refunded
	DisputeReason  string     `json:"dispute_reason"`
	RefundAmount   float64    `json:"refund_amount"`
	RefundedAt     *time.Time `json:"refunded_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

