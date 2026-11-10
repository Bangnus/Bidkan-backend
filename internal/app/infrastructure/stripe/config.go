package stripe

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAPIBaseURL         = "https://api.stripe.com"
	DefaultConnectCountry     = "TH"
	DefaultConnectAccountType = "express"
	DefaultBackofficeURL      = "http://localhost:3001"
	DefaultClientURL          = "http://localhost:3000"
	DefaultAutoPayoutInterval = 24 * time.Hour
)

type Config struct {
	SecretKey            string
	WebhookSecret        string
	ConnectWebhookSecret string
	ConnectAccountType   string
	ConnectCountry       string
	BackofficeURL        string
	ClientURL            string
	APIBaseURL           string
	AutoPayoutEnabled    bool
	AutoPayoutInterval   time.Duration
}

func LoadConfigFromEnv() *Config {
	accountType := strings.ToLower(strings.TrimSpace(os.Getenv("STRIPE_CONNECT_ACCOUNT_TYPE")))
	if accountType != "standard" {
		accountType = DefaultConnectAccountType
	}
	country := strings.ToUpper(strings.TrimSpace(os.Getenv("STRIPE_CONNECT_COUNTRY")))
	if country == "" {
		country = DefaultConnectCountry
	}
	backofficeURL := strings.TrimRight(strings.TrimSpace(os.Getenv("STRIPE_BACKOFFICE_URL")), "/")
	if backofficeURL == "" {
		backofficeURL = DefaultBackofficeURL
	}
	clientURL := strings.TrimRight(strings.TrimSpace(os.Getenv("STRIPE_CLIENT_URL")), "/")
	if clientURL == "" {
		clientURL = DefaultClientURL
	}
	apiBase := strings.TrimRight(strings.TrimSpace(os.Getenv("STRIPE_API_BASE")), "/")
	if apiBase == "" {
		apiBase = DefaultAPIBaseURL
	}
	autoEnabled, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("STRIPE_AUTO_PAYOUT_ENABLED")))
	interval := DefaultAutoPayoutInterval
	if raw := strings.TrimSpace(os.Getenv("STRIPE_AUTO_PAYOUT_INTERVAL")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed >= time.Minute {
			interval = parsed
		}
	}

	return &Config{
		SecretKey:            strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")),
		WebhookSecret:        strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET")),
		ConnectWebhookSecret: strings.TrimSpace(os.Getenv("STRIPE_CONNECT_WEBHOOK_SECRET")),
		ConnectAccountType:   accountType,
		ConnectCountry:       country,
		BackofficeURL:        backofficeURL,
		ClientURL:            clientURL,
		APIBaseURL:           apiBase,
		AutoPayoutEnabled:    autoEnabled,
		AutoPayoutInterval:   interval,
	}
}

func (c *Config) IsConfigured() bool {
	return c != nil && c.SecretKey != ""
}

func (c *Config) WebhookSecrets() []string {
	if c == nil {
		return nil
	}
	var out []string
	seen := map[string]bool{}
	for _, s := range []string{c.WebhookSecret, c.ConnectWebhookSecret} {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

