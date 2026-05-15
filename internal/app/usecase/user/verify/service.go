package verify

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Request struct {
	OTP string `json:"otp" validate:"required,len=6" example:"123456"`
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
	// 1. ตรวจสอบเบอร์โทรจากรหัส OTP ที่ส่งมา
	phone, err := s.otpRepo.VerifyOTP(ctx, req.OTP)
	if err != nil || phone == "" {
		return errors.New("invalid or expired OTP")
	}

	// 2. ดึงข้อมูลผู้ใช้จากเบอร์โทรที่ได้จาก OTP
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return errors.New("user not found for this OTP")
	}

	// 3. อัปเดตสถานะเป็น active
	err = s.userRepo.UpdateStatus(ctx, user.ID, "active")
	if err != nil {
		return errors.New("failed to activate user")
	}

	// 4. ลบ OTP ออกจาก Redis
	_ = s.otpRepo.DeleteOTP(ctx, req.OTP)

	return nil
}
