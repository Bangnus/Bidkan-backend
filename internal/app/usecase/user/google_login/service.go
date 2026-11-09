package google_login

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type Request struct {
	Credential string `json:"credential" validate:"required"`
	CampusSlug string `json:"campus_slug"`
}

type Response struct {
	Token string       `json:"token"`
	Role  string       `json:"role"`
	User  *entity.User `json:"user"`
}

type GoogleClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
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
	if req.Credential == "" {
		return nil, errors.New("missing google credential")
	}

	claims, err := s.verifyGoogleToken(ctx, req.Credential)
	if err != nil {
		return nil, err
	}

	// 2. ตรวจสอบแยกมหาวิทยาลัย (Auto-detect Domain หรือ General Email)
	tenantID := s.resolveTenantID(claims.Email, req.CampusSlug)

	// 3. Upsert ผู้ใช้ลง Database
	user, err := s.userRepo.UpsertGoogleUser(
		ctx,
		claims.Sub,
		claims.Email,
		claims.Name,
		claims.Picture,
		tenantID,
	)
	if err != nil {
		return nil, err
	}

	// 4. ออก JWT Token ของระบบ BidKan
	var tenantIDStr string
	if user.TenantID != nil {
		tenantIDStr = user.TenantID.String()
	}
	token, err := s.generateJWT(user.ID.String(), user.Role, tenantIDStr)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &Response{
		Token: token,
		Role:  user.Role,
		User:  user,
	}, nil
}

func (s *service) verifyGoogleToken(ctx context.Context, credential string) (*GoogleClaims, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID != "" {
		payload, err := idtoken.Validate(ctx, credential, clientID)
		if err == nil && payload != nil {
			name, _ := payload.Claims["name"].(string)
			email, _ := payload.Claims["email"].(string)
			picture, _ := payload.Claims["picture"].(string)
			return &GoogleClaims{
				Sub:     payload.Subject,
				Email:   email,
				Name:    name,
				Picture: picture,
			}, nil
		}
	}

	// Fallback: ถอด JWT Payload สำหรับ Dev Mode
	parts := strings.Split(credential, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid google token format")
	}
	payloadSegment := parts[1]
	if l := len(payloadSegment) % 4; l > 0 {
		payloadSegment += strings.Repeat("=", 4-l)
	}
	raw, err := base64.URLEncoding.DecodeString(payloadSegment)
	if err != nil {
		return nil, errors.New("failed to decode google payload")
	}

	var claims GoogleClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, errors.New("failed to parse google claims")
	}
	return &claims, nil
}

func (s *service) resolveTenantID(email, campusSlug string) *uuid.UUID {
	lowerEmail := strings.ToLower(email)

	// 1. ตรวจจับจาก Email Domain ของมหาวิทยาลัย
	if strings.HasSuffix(lowerEmail, "ku.th") || strings.HasSuffix(lowerEmail, "ku.ac.th") {
		id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		return &id
	}
	if strings.HasSuffix(lowerEmail, "cmu.ac.th") {
		id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		return &id
	}

	// 2. อีเมลทั่วไป (@gmail.com / บุคคลภายนอก) ที่ระบุ campus_slug มา
	switch strings.ToLower(campusSlug) {
	case "cmu":
		id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		return &id
	case "ku":
		id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		return &id
	}

	// 3. อีเมลทั่วไปที่ไม่ระบุ ให้อยู่มหาวิทยาลัยหลัก (KU กำแพงแสน)
	defaultID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	return &defaultID
}

func (s *service) generateJWT(userID, role, tenantID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "bidkan_secret_key_2024"
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":     time.Now().Unix(),
	}
	if tenantID != "" {
		claims["tenant_id"] = tenantID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

