package create

import (
	"context"
	"errors"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"

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
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{
		userRepo:    userRepo,
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

	// 3. เตรียมข้อมูล Entity (สถานะเป็น pending เพื่อรอการยืนยันจาก Firebase)
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

	// ไม่ต้องส่ง OTP เองแล้ว ให้ Frontend จัดการผ่าน Firebase SDK

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
