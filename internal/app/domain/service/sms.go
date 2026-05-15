package service

import "context"

type SmsProvider interface {
	SendOTP(ctx context.Context, phone string, otp string) error
}
