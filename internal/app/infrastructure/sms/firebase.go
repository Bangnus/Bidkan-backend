package sms

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"google.golang.org/api/option"
)

type firebaseSmsProvider struct {
	app        *firebase.App
	authClient *auth.Client
}

func NewFirebaseSmsProvider(credentialsFile string) (service.SmsProvider, error) {
	// 1. ตั้งค่า Credentials จากไฟล์ JSON
	opt := option.WithCredentialsFile(credentialsFile)
	
	// 2. เริ่มต้น Firebase App
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	// 3. เริ่มต้น Auth Client (เผื่อใช้จัดการเรื่อง Phone Verification)
	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting firebase auth client: %v", err)
	}

	return &firebaseSmsProvider{
		app:        app,
		authClient: authClient,
	}, nil
}

func (p *firebaseSmsProvider) SendOTP(ctx context.Context, phone string, otp string) error {
	// แม้จะใช้ Firebase Verify จาก Frontend แต่เราต้องมีฟังก์ชันนี้ไว้ตาม Interface ครับ
	fmt.Printf("[FIREBASE SDK] SendOTP called for %s (Mocking - No real SMS sent from backend)\n", phone)
	return nil
}

func (p *firebaseSmsProvider) VerifyIDToken(ctx context.Context, idToken string) (string, error) {
	// --- DEBUG BYPASS ---
	// ถ้าส่งรหัสมาในรูปแบบ debug:0812345678 ให้ถือว่าผ่านทันที (สำหรับ Dev)
	if len(idToken) > 6 && idToken[:6] == "debug:" {
		phone := idToken[6:]
		fmt.Printf("[FIREBASE DEBUG] 🔓 Bypassing verification for phone: %s\n", phone)
		return phone, nil
	}
	// --------------------

	token, err := p.authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}

	// ดึงเบอร์โทรศัพท์ออกมาจาก Token
	phoneNumber, ok := token.Claims["phone_number"].(string)
	if !ok {
		return "", fmt.Errorf("phone number not found in token")
	}

	return phoneNumber, nil
}
