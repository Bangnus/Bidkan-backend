package sms

import (
	"context"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
)

type firebaseSmsProvider struct {
	// ในอนาคตจะใส่ Firebase Client ตรงนี้
	apiKey string 
}

func NewFirebaseSmsProvider(apiKey string) service.SmsProvider {
	return &firebaseSmsProvider{apiKey: apiKey}
}

func (p *firebaseSmsProvider) SendOTP(ctx context.Context, phone string, otp string) error {
	// TODO: เชื่อมต่อ Firebase Admin SDK หรือ Firebase Auth REST API
	fmt.Printf("[FIREBASE SMS] Sending to %s via Firebase... (Mocking API Call)\n", phone)
	return nil
}
