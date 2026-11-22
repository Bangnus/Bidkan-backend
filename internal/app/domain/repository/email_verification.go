package repository

import (
	"context"
	"time"
)

// EmailVerificationRepository stores temporary verification codes for email signups.
type EmailVerificationRepository interface {
	SaveCode(ctx context.Context, email, code string, ttl time.Duration) error
	VerifyCode(ctx context.Context, email, code string, maxAttempts int) (bool, error)
	ResendCooldownRemaining(ctx context.Context, email string) (time.Duration, error)
	SetResendCooldown(ctx context.Context, email string, ttl time.Duration) error
}
