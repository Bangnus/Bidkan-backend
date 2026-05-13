package create

import (
	"context"
	"errors"
	"time"

	"bidkan/internal/app/domain/entity"
	"bidkan/internal/app/domain/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Request Data (รับมาจาก Client)
type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=6"`
}

// Response Data (ส่งกลับให้ Client - ห้ามส่ง Password กลับเด็ดขาด)
type Response struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
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
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
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
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hashedPassword),
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
		Email: user.Email,
		Name:  user.Name,
	}, nil
}