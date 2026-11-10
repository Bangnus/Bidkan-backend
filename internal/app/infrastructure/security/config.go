package security

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"
)

const (
	DefaultAccessTTL  = 30 * time.Minute
	DefaultRefreshTTL = 7 * 24 * time.Hour
	MinSecretLength   = 32
	legacyWeakSecret  = "bidkan_secret_key_2024"
)

type Config struct {
	JWTSecret          []byte
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	AppEnv             string
	CORSAllowedOrigins string
	GoogleClientID     string
	EnableSwagger      bool
}

func LoadConfig() (*Config, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}
	if secret == legacyWeakSecret || len(secret) < MinSecretLength {
		return nil, errors.New("JWT_SECRET must be at least 32 characters and not use default fallback")
	}

	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if appEnv == "" {
		appEnv = "development"
	}

	corsOrigins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if corsOrigins == "" {
		corsOrigins = "http://localhost:3000,http://localhost:3001"
	}

	accessTTL := parseDurationEnv("JWT_ACCESS_TTL", DefaultAccessTTL)
	refreshTTL := parseDurationEnv("JWT_REFRESH_TTL", DefaultRefreshTTL)

	enableSwagger := appEnv != "production"
	if strings.EqualFold(os.Getenv("ENABLE_SWAGGER"), "false") {
		enableSwagger = false
	}

	return &Config{
		JWTSecret:                 []byte(secret),
		AccessTokenTTL:            accessTTL,
		RefreshTokenTTL:           refreshTTL,
		AppEnv:             appEnv,
		CORSAllowedOrigins: corsOrigins,
		GoogleClientID:     strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		EnableSwagger:      enableSwagger,
	}, nil
}

func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Fatal security configuration error: %v", err)
	}
	return cfg
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

