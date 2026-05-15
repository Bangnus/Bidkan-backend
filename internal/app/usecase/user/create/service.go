package create

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Request Data
type Request struct {
	PhoneNumber string `json:"phone_number" validate:"required,min=10" example:"0812345678"`
	Username    string `json:"username" validate:"required,min=2" example:"somchai_jaidee"`
	Password    string `json:"password" validate:"required,min=6" example:"password123"`
}

// Response Data
type Response struct {
	ID            string    `json:"id"`
	PhoneNumber   string    `json:"phone_number"`
	Username      string    `json:"username"`
	WalletBalance string    `json:"wallet_balance"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Service interface {
	CreateUser(ctx context.Context, req Request) (*Response, error)
}

type service struct {
	userRepo    repository.UserRepository
	otpRepo     repository.OtpRepository
	smsProvider domainService.SmsProvider
}

func NewService(userRepo repository.UserRepository, otpRepo repository.OtpRepository, smsProvider domainService.SmsProvider) Service {
	return &service{
		userRepo:    userRepo,
		otpRepo:     otpRepo,
		smsProvider: smsProvider,
	}
}

func (s *service) CreateUser(ctx context.Context, req Request) (*Response, error) {
	// 1. ตรวจสอบว่าเบอร์โทรซ้ำหรือไม่
	existing, _ := s.userRepo.GetByPhone(ctx, req.PhoneNumber)
	if existing != nil {
		return nil, errors.New("phone number already registered")
	}

	// 2. เข้ารหัสผ่านด้วย Bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 3. เตรียมข้อมูล Entity (สถานะเป็น pending)
	now := time.Now()
	user := &entity.User{
		ID:            uuid.New(),
		PhoneNumber:   req.PhoneNumber,
		Username:      req.Username,
		Password:      string(hashedPassword),
		WalletBalance: "0.00",
		Role:          "user",
		Status:        "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 4. บันทึกลงฐานข้อมูล
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("failed to create user in database")
	}

	// 5. เจนรหัส OTP และบันทึกลง Redis
	otpCode := generateRandomOTP(6)
	err = s.otpRepo.SaveOTP(ctx, user.PhoneNumber, otpCode, 5*time.Minute)
	if err != nil {
		return nil, errors.New("user created but failed to send OTP")
	}

	// 6. ส่ง SMS ผ่าน Provider
	_ = s.smsProvider.SendOTP(ctx, user.PhoneNumber, otpCode)

	return &Response{
		ID:            user.ID.String(),
		PhoneNumber:   user.PhoneNumber,
		Username:      user.Username,
		WalletBalance: user.WalletBalance,
		Role:          user.Role,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

func generateRandomOTP(length int) string {
	table := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, length)
	n, err := io.ReadAtLeast(rand.Reader, b, length)
	if n != length || err != nil {
		return "123456"
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}
