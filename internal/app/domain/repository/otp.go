package repository

import (
	"context"
	"time"
)

type OtpRepository interface {
	SaveOTP(ctx context.Context, phone string, otp string, expiration time.Duration) error
	VerifyOTP(ctx context.Context, phone string, otp string) (bool, error)
	DeleteOTP(ctx context.Context, phone string) error
}
