package stripe_topup

import (
	"context"

	"github.com/google/uuid"
)

const (
	MinTopupTHB = 10.0
	MaxTopupTHB = 50000.0
)

type ServiceError struct {
	Status int
	Code   string
	Msg    string
}

func (e *ServiceError) Error() string { return e.Msg }

type ConfigResponse struct {
	StripeEnabled bool    `json:"stripe_enabled"`
	Currency      string  `json:"currency"`
	MinAmount     float64 `json:"min_amount"`
	MaxAmount     float64 `json:"max_amount"`
}

type CheckoutRequest struct {
	UserID uuid.UUID `json:"-"`
	Amount float64   `json:"amount"`
	Method string    `json:"method"` // "card" | "promptpay" | ""
}

type CheckoutResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	SessionID     string    `json:"session_id"`
	CheckoutURL   string    `json:"checkout_url"`
	Amount        float64   `json:"amount"`
}

type VerifyResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	SessionID     string    `json:"session_id"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	Credited      bool      `json:"credited"`
}

type Service interface {
	GetConfig() ConfigResponse
	CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)
	VerifySession(ctx context.Context, userID uuid.UUID, sessionID string) (*VerifyResponse, error)
}

