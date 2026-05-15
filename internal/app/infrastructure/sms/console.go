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
func (p *consoleSmsProvider) VerifyIDToken(ctx context.Context, idToken string) (string, error) {
	// ท่าไม้ตายสำหรับ Dev: ถ้าส่งมาในรูปแบบ debug:0812345678
	if len(idToken) > 6 && idToken[:6] == "debug:" {
		phone := idToken[6:]
		fmt.Printf("[DEV MODE] 🔓 Bypass Firebase for phone: %s\n", phone)
		return phone, nil
	}

	return "", fmt.Errorf("invalid console token (use debug:08xxxxxxxx for testing)")
}
