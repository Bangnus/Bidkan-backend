package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Request struct {
	PhoneNumber string `json:"phone_number" validate:"required,min=10" example:"0812345678"`
}

type Service interface {
	SendOTP(ctx context.Context, req Request) error
}

type service struct {
	otpRepo repository.OtpRepository
}

func NewService(otpRepo repository.OtpRepository) Service {
	return &service{otpRepo: otpRepo}
}

func (s *service) SendOTP(ctx context.Context, req Request) error {
	// 1. Generate OTP (สุ่มเลข 6 หลัก)
	otp := encodeCursor(6)
	
	// 2. Save to Redis (หมดอายุใน 5 นาที)
	err := s.otpRepo.SaveOTP(ctx, req.PhoneNumber, otp, 5*time.Minute)
	if err != nil {
		return err
	}

	// 3. TODO: ส่ง SMS ผ่าน Gateway จริง
	fmt.Printf("Sending OTP %s to %s\n", otp, req.PhoneNumber)
	
	return nil
}

// ฟังก์ชันช่วยสุ่มตัวเลข
func encodeCursor(length int) string {
	max := length
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max || err != nil {
		return "123456" // fallback
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
