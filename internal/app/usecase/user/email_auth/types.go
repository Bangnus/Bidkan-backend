package email_auth

import (
	"errors"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
)

const (
	VerificationCodeTTL = 15 * time.Minute
	ResendCooldown      = 60 * time.Second
	MaxVerifyAttempts   = 5
)

var (
	ErrEmailExists          = errors.New("อีเมลนี้ถูกสมัครใช้งานแล้ว กรุณาเข้าสู่ระบบ")
	ErrPhoneExists          = errors.New("เบอร์โทรศัพท์นี้ถูกใช้งานแล้ว")
	ErrAccountRestricted    = errors.New("บัญชีนี้ถูกระงับการใช้งาน")
	ErrAlreadyVerified      = errors.New("อีเมลนี้ได้รับการยืนยันแล้ว กรุณาเข้าสู่ระบบ")
	ErrInvalidVerification  = errors.New("รหัสยืนยันไม่ถูกต้องหรือหมดอายุแล้ว")
	ErrResendTooSoon        = errors.New("กรุณารอสักครู่ก่อนขอรหัสยืนยันใหม่อีกครั้ง")
	ErrPendingAccountNotSet = errors.New("ไม่พบบัญชีที่รอยืนยันสำหรับอีเมลนี้")
)

type CooldownError struct {
	RetryAfterSeconds int
}

func (e *CooldownError) Error() string { return ErrResendTooSoon.Error() }
func (e *CooldownError) Unwrap() error { return ErrResendTooSoon }

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email,max=254"`
	Username    string `json:"username" validate:"required,min=2,max=50"`
	Password    string `json:"password" validate:"required,min=6,max=72"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,min=9,max=20"`
	CampusSlug  string `json:"campus_slug"`
}

type VerificationDispatchResponse struct {
	Message               string `json:"message"`
	Email                 string `json:"email"`
	ExpiresInSeconds      int    `json:"expires_in_seconds"`
	ResendCooldownSeconds int    `json:"resend_cooldown_seconds"`
	DeliveryMode          string `json:"delivery_mode"`
	DevOTP                string `json:"dev_otp,omitempty"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

type VerifyEmailResponse struct {
	Token        string       `json:"token"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	Role         string       `json:"role"`
	User         *entity.User `json:"user"`
}

type ResendRequest struct {
	Email string `json:"email" validate:"required,email"`
}
