package email_auth

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*VerificationDispatchResponse, error)
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) (*VerifyEmailResponse, error)
	ResendVerification(ctx context.Context, req ResendRequest) (*VerificationDispatchResponse, error)
}

type service struct {
	userRepo      repository.UserRepository
	tenantRepo    repository.TenantRepository
	verifyRepo    repository.EmailVerificationRepository
	emailProvider domainService.EmailProvider
	secCfg        *security.Config
	sessions      security.SessionStore
}

func NewService(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	verifyRepo repository.EmailVerificationRepository,
	emailProvider domainService.EmailProvider,
	secCfg *security.Config,
	sessions security.SessionStore,
) Service {
	return &service{
		userRepo:      userRepo,
		tenantRepo:    tenantRepo,
		verifyRepo:    verifyRepo,
		emailProvider: emailProvider,
		secCfg:        secCfg,
		sessions:      sessions,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*VerificationDispatchResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	username := strings.TrimSpace(req.Username)
	phone := strings.TrimSpace(req.PhoneNumber)
	if strings.Contains(username, "@") {
		return nil, ErrInvalidUsername
	}

	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		if existing.Status == "suspended" {
			return nil, ErrAccountRestricted
		}
		if existing.Status == "active" || existing.EmailVerified || existing.Password == "google_oauth" {
			return nil, ErrEmailExists
		}
		if rem, err := s.verifyRepo.ResendCooldownRemaining(ctx, email); err == nil && rem > 0 {
			return nil, &CooldownError{RetryAfterSeconds: int(math.Ceil(rem.Seconds()))}
		}
	}

	if byUser, _ := s.userRepo.GetByUsername(ctx, username); byUser != nil && (existing == nil || byUser.ID != existing.ID) {
		return nil, ErrUsernameTaken
	}

	if phone != "" {
		if byPhone, _ := s.userRepo.GetByPhone(ctx, phone); byPhone != nil && (existing == nil || byPhone.ID != existing.ID) {
			return nil, ErrPhoneExists
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	tenantID := resolveTenantID(ctx, s.tenantRepo, email, req.CampusSlug)
	if existing != nil {
		existing.Username = username
		existing.PhoneNumber = phone
		existing.Password = string(hashedPassword)
		existing.TenantID = tenantID
		if _, err := s.userRepo.RefreshPendingEmailUser(ctx, existing); err != nil {
			return nil, errors.New("failed to update pending registration")
		}
	} else {
		_, err = s.userRepo.CreateEmailUser(ctx, &entity.User{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Email:       email,
			Username:    username,
			PhoneNumber: phone,
			Password:    string(hashedPassword),
		})
		if err != nil {
			return nil, errors.New("failed to create user account")
		}
	}

	return s.dispatchOTP(ctx, email, username)
}

func (s *service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) (*VerifyEmailResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, ErrInvalidVerification
	}
	if user.Status == "suspended" {
		return nil, ErrAccountRestricted
	}
	if user.Status == "active" && user.EmailVerified {
		return nil, ErrAlreadyVerified
	}

	ok, err := s.verifyRepo.VerifyCode(ctx, email, code, MaxVerifyAttempts)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidVerification
	}

	activated, err := s.userRepo.ActivateVerifiedEmailUser(ctx, email)
	if err != nil {
		return nil, errors.New("failed to activate account")
	}
	return s.issueSessionResponse(ctx, activated)
}

func (s *service) ResendVerification(ctx context.Context, req ResendRequest) (*VerificationDispatchResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, ErrPendingAccountNotSet
	}
	if user.Status == "suspended" {
		return nil, ErrAccountRestricted
	}
	if user.Status == "active" && user.EmailVerified {
		return nil, ErrAlreadyVerified
	}
	if rem, err := s.verifyRepo.ResendCooldownRemaining(ctx, email); err == nil && rem > 0 {
		return nil, &CooldownError{RetryAfterSeconds: int(math.Ceil(rem.Seconds()))}
	}
	return s.dispatchOTP(ctx, email, user.Username)
}

func (s *service) dispatchOTP(ctx context.Context, email, username string) (*VerificationDispatchResponse, error) {
	otp, err := generateSixDigitOTP()
	if err != nil {
		return nil, errors.New("failed to generate verification code")
	}
	if err := s.verifyRepo.SaveCode(ctx, email, otp, VerificationCodeTTL); err != nil {
		return nil, errors.New("failed to store verification code")
	}

	msg := buildVerificationEmail(email, username, otp)
	if err := s.emailProvider.Send(ctx, msg); err != nil {
		return nil, errors.New("ส่งอีเมลยืนยันไม่สำเร็จ กรุณาตรวจสอบการตั้งค่า SMTP หรือลองใหม่อีกครั้ง")
	}
	_ = s.verifyRepo.SetResendCooldown(ctx, email, ResendCooldown)

	resp := &VerificationDispatchResponse{
		Message:               "ส่งรหัสยืนยัน 6 หลักไปยังอีเมลของคุณแล้ว",
		Email:                 email,
		ExpiresInSeconds:      int(VerificationCodeTTL.Seconds()),
		ResendCooldownSeconds: int(ResendCooldown.Seconds()),
		DeliveryMode:          s.emailProvider.Mode(),
	}
	if s.secCfg != nil && s.secCfg.AppEnv == "development" && s.emailProvider.Mode() == "console" {
		resp.DevOTP = otp
	}
	return resp, nil
}
