package verify

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Request struct {
	PhoneNumber string `json:"phone_number" validate:"required,min=10" example:"0812345678"`
	OTP         string `json:"otp" validate:"required,len=6" example:"123456"`
}

type Service interface {
	VerifyOTP(ctx context.Context, req Request) error
}

type service struct {
	userRepo repository.UserRepository
	otpRepo  repository.OtpRepository
}

func NewService(userRepo repository.UserRepository, otpRepo repository.OtpRepository) Service {
	return &service{
		userRepo: userRepo,
		otpRepo:  otpRepo,
	}
}

func (s *service) VerifyOTP(ctx context.Context, req Request) error {
	// 1. ตรวจสอบ OTP จาก Redis
	isValid, err := s.otpRepo.VerifyOTP(ctx, req.PhoneNumber, req.OTP)
	if err != nil || !isValid {
		return errors.New("invalid or expired OTP")
	}

	// 2. ดึงข้อมูลผู้ใช้จากเบอร์โทร
	user, err := s.userRepo.GetByPhone(ctx, req.PhoneNumber)
	if err != nil {
		return errors.New("user not found")
	}

	// 3. อัปเดตสถานะเป็น active
	err = s.userRepo.UpdateStatus(ctx, user.ID, "active")
	if err != nil {
		return errors.New("failed to activate user")
	}

	// 4. ลบ OTP ออกจาก Redis
	_ = s.otpRepo.DeleteOTP(ctx, req.PhoneNumber)

	return nil
}
