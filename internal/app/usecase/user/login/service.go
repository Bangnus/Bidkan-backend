package login

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Request struct {
	Username string `json:"username" validate:"required" example:"somchai_jaidee"`
	Password string `json:"password" validate:"required" example:"password123"`
}

type Response struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

type Service interface {
	Login(ctx context.Context, req Request) (*Response, error)
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) Login(ctx context.Context, req Request) (*Response, error) {
	// 1. ค้นหาผู้ใช้จาก Username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 2. ตรวจสอบสถานะ (ต้องเป็น active เท่านั้น)
	if user.Status != "active" {
		return nil, errors.New("account is not activated. please verify your OTP first")
	}

	// 3. ตรวจสอบรหัสผ่าน (Bcrypt)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 4. สร้าง JWT Token
	token, err := s.generateJWT(user.ID.String(), user.Role)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &Response{
		Token: token,
		Role:  user.Role,
	}, nil
}

func (s *service) generateJWT(userID string, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "bidkan_secret_key_2024" // Default สำหรับ Dev
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // อายุ 7 วัน
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
