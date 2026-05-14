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

// Request Data (รับมาจาก Client)
type Request struct {
	PhoneNumber    string `json:"phone_number" validate:"required,min=10"`
	FullName     string `json:"full_name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=6"`
}

// Response Data (ส่งกลับให้ Client - ห้ามส่ง Password กลับเด็ดขาด)
type Response struct {
	ID            string    `json:"id"`
	PhoneNumber   string    `json:"phone_number"`
	FullName      string    `json:"full_name"`
	WalletBalance string    `json:"wallet_balance"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Interface สำหรับ Mock ในการทำ Unit Test
type Service interface {
	CreateUser(ctx context.Context, req Request) (*Response, error)
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) CreateUser(ctx context.Context, req Request) (*Response, error) {
	// 1. ตรวจสอบว่าอีเมลซ้ำหรือไม่
	existing, _ := s.userRepo.GetByPhone(ctx, req.PhoneNumber)
	if existing != nil {
		return nil, errors.New("phone number already registered")
	}

	// 2. เข้ารหัสผ่านด้วย Bcrypt (Cost 12)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 3. เตรียมข้อมูล Entity
	now := time.Now()
	user := &entity.User{
		ID:        uuid.New(),
		PhoneNumber:    req.PhoneNumber,
		FullName:     req.FullName,
		Password:  string(hashedPassword),
		WalletBalance: "0.00",
		Role:        "user",
		Status:      "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 4. บันทึกลงฐานข้อมูล
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("failed to create user in database")
	}

	// 5. คืนค่า Response ออกไป
	return &Response{
		ID:    user.ID.String(),
		PhoneNumber: user.PhoneNumber,
		FullName:  user.FullName,
		WalletBalance: user.WalletBalance,
		Role: user.Role,
		Status: user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}