package verify_firebase

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
)

type Request struct {
	IdToken string `json:"id_token" validate:"required"`
}

type Service interface {
	VerifyAndActivate(ctx context.Context, req Request) error
}

type service struct {
	userRepo    repository.UserRepository
	smsProvider domainService.SmsProvider
}

func NewService(userRepo repository.UserRepository, smsProvider domainService.SmsProvider) Service {
	return &service{
		userRepo:    userRepo,
		smsProvider: smsProvider,
	}
}

func (s *service) VerifyAndActivate(ctx context.Context, req Request) error {
	// 1. ตรวจสอบ IdToken กับ Firebase
	phone, err := s.smsProvider.VerifyIDToken(ctx, req.IdToken)
	if err != nil {
		return errors.New("invalid or expired firebase token")
	}

	// 2. ดึงข้อมูลผู้ใช้จากเบอร์โทรที่ได้มาจาก Firebase
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return errors.New("user not found for this phone number")
	}

	// 3. อัปเดตสถานะเป็น active
	err = s.userRepo.UpdateStatus(ctx, user.ID, "active")
	if err != nil {
		return errors.New("failed to activate user")
	}

	return nil
}
