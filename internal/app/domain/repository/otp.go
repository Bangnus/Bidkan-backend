package repository

import (
	"context"
	"time"
)

type OtpRepository interface {
	SaveOTP(ctx context.Context, phone string, otp string, expiration time.Duration) error
	VerifyOTP(ctx context.Context, otp string) (string, error)
	DeleteOTP(ctx context.Context, otp string) error
}
