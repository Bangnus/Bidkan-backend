package sms

import (
	"context"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
)

type consoleSmsProvider struct{}

func NewConsoleSmsProvider() service.SmsProvider {
	return &consoleSmsProvider{}
}

func (p *consoleSmsProvider) SendOTP(ctx context.Context, phone string, otp string) error {
	fmt.Println("-----------------------------------------")
	fmt.Printf("[CONSOLE SMS] To: %s | Message: Your OTP is %s\n", phone, otp)
	fmt.Println("-----------------------------------------")
	return nil
}
