package otp

import (
	"context"
	"crypto/rand"
	"io"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
)

type Request struct {
	PhoneNumber string `json:"phone_number" validate:"required,min=10" example:"0812345678"`
}

type Service interface {
	SendOTP(ctx context.Context, req Request) error
}

type service struct {
	otpRepo     repository.OtpRepository
	smsProvider domainService.SmsProvider
}

func NewService(otpRepo repository.OtpRepository, smsProvider domainService.SmsProvider) Service {
	return &service{
		otpRepo:     otpRepo,
		smsProvider: smsProvider,
	}
}

func (s *service) SendOTP(ctx context.Context, req Request) error {
	// 1. Generate OTP (สุ่มเลข 6 หลัก)
	otp := encodeCursor(6)
	
	// 2. Save to Redis (หมดอายุใน 5 นาที)
	err := s.otpRepo.SaveOTP(ctx, req.PhoneNumber, otp, 5*time.Minute)
	if err != nil {
		return err
	}

	// 3. ส่ง SMS ผ่าน Provider 
	return s.smsProvider.SendOTP(ctx, req.PhoneNumber, otp)
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func encodeCursor(length int) string {
	max := length
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max || err != nil {
		return "123456" 
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}
