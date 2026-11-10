package stripe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrNotConfigured = errors.New("stripe is not configured")

type APIError struct {
	StatusCode int
	Type       string `json:"type"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("stripe (%s): %s", e.Code, e.Message)
	}
	return fmt.Sprintf("stripe (%d): %s", e.StatusCode, e.Message)
}

func IsMissingResource(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound || apiErr.Code == "resource_missing"
	}
	return false
}

type AccountStatus struct {
	ID               string            `json:"id"`
	ChargesEnabled   bool              `json:"charges_enabled"`
	PayoutsEnabled   bool              `json:"payouts_enabled"`
	DetailsSubmitted bool              `json:"details_submitted"`
	CurrentlyDue     []string          `json:"currently_due"`
	DisabledReason   string            `json:"disabled_reason"`
	Metadata         map[string]string `json:"metadata"`
}

type TransferResult struct {
	ID          string `json:"id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Destination string `json:"destination"`
	Reversed    bool   `json:"reversed"`
}

type CheckoutSessionParams struct {
	TransactionID string
	UserID        string
	AmountTHB     float64
	CustomerEmail string
	Method        string // "card", "promptpay", or "" (both)
	SuccessURL    string
	CancelURL     string
}

type CheckoutSession struct {
	ID                string            `json:"id"`
	URL               string            `json:"url"`
	Status            string            `json:"status"`
	PaymentStatus     string            `json:"payment_status"`
	AmountTotal       int64             `json:"amount_total"`
	Currency          string            `json:"currency"`
	ClientReferenceID string            `json:"client_reference_id"`
	PaymentIntent     *string           `json:"payment_intent"`
	Metadata          map[string]string `json:"metadata"`
}

type Client interface {
	IsConfigured() bool
	Config() *Config
	CreateConnectedAccount(ctx context.Context, slug, name, email string) (string, error)
	GetAccount(ctx context.Context, accountID string) (*AccountStatus, error)
	CreateAccountLink(ctx context.Context, accountID, refreshURL, returnURL string) (string, error)
	CreateExpressLoginLink(ctx context.Context, accountID string) (string, error)
	CreateTransfer(ctx context.Context, destinationAccount string, amountTHB float64, idempotencyKey string, metadata map[string]string) (*TransferResult, error)
	CreateCheckoutSession(ctx context.Context, params CheckoutSessionParams) (*CheckoutSession, error)
	GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error)
}

type httpClient struct {
	cfg  *Config
	http *http.Client
}

func NewClient(cfg *Config) Client {
	if cfg == nil {
		cfg = LoadConfigFromEnv()
	}
	return &httpClient{
		cfg:  cfg,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *httpClient) IsConfigured() bool { return c != nil && c.cfg.IsConfigured() }
func (c *httpClient) Config() *Config    { return c.cfg }

func (c *httpClient) doRequest(ctx context.Context, method, path string, form url.Values, idempotencyKey string, out any) error {
	if !c.IsConfigured() {
		return ErrNotConfigured
	}
	var bodyReader io.Reader
	if form != nil {
		bodyReader = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.APIBaseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.cfg.SecretKey, "")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var envelope struct {
			Error APIError `json:"error"`
		}
		if jsonErr := json.Unmarshal(raw, &envelope); jsonErr == nil && envelope.Error.Message != "" {
			envelope.Error.StatusCode = resp.StatusCode
			return &envelope.Error
		}
		return &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(raw))}
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

