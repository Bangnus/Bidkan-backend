package login

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrTooManyAttempts   = errors.New("account temporarily locked due to failed login attempts")
	ErrForbiddenScope    = errors.New("admin accounts must sign in via backoffice portal")
	ErrMissingIdentifier = errors.New("email or username is required")
	ErrEmailNotVerified  = errors.New("อีเมลยังไม่ได้รับการยืนยัน กรุณายืนยันรหัส OTP ก่อนเข้าสู่ระบบ")
)

type EmailNotVerifiedError struct {
	Email string
}

func (e *EmailNotVerifiedError) Error() string { return ErrEmailNotVerified.Error() }
func (e *EmailNotVerifiedError) Unwrap() error { return ErrEmailNotVerified }

type Request struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Password   string `json:"password" validate:"required" example:"password123"`
}

type Response struct {
	Token        string       `json:"token"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	Role         string       `json:"role"`
	User         *entity.User `json:"user,omitempty"`
}

type Service interface {
	Login(ctx context.Context, req Request) (*Response, error)
}

type service struct {
	userRepo repository.UserRepository
	secCfg   *security.Config
	sessions security.SessionStore
}

func NewService(userRepo repository.UserRepository, secCfg *security.Config, sessions security.SessionStore) Service {
	return &service{userRepo: userRepo, secCfg: secCfg, sessions: sessions}
}

func (s *service) Login(ctx context.Context, req Request) (*Response, error) {
	identifier := firstNonEmpty(req.Identifier, req.Email, req.Username)
	if identifier == "" {
		return nil, ErrMissingIdentifier
	}
	lockKey := "rider:" + strings.ToLower(identifier)
	if s.sessions != nil {
		if locked, _, _ := s.sessions.IsLoginLocked(ctx, lockKey); locked {
			return nil, ErrTooManyAttempts
		}
	}

	user, err := s.userRepo.GetByIdentifier(ctx, identifier)
	if err != nil || user == nil {
		s.recordFailure(ctx, lockKey)
		return nil, errors.New("อีเมล/ชื่อผู้ใช้ หรือรหัสผ่านไม่ถูกต้อง")
	}

	if user.Password == "" || user.Password == "google_oauth" {
		s.recordFailure(ctx, lockKey)
		return nil, errors.New("บัญชีนี้ลงทะเบียนผ่าน Google กรุณาเข้าสู่ระบบด้วยปุ่ม Google")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.recordFailure(ctx, lockKey)
		return nil, errors.New("อีเมล/ชื่อผู้ใช้ หรือรหัสผ่านไม่ถูกต้อง")
	}

	if user.Role != "user" {
		return nil, ErrForbiddenScope
	}

	if user.Status == "pending" || (user.Email != "" && !user.EmailVerified && user.Status != "suspended") {
		return nil, &EmailNotVerifiedError{Email: user.Email}
	}

	if user.Status != "active" {
		return nil, errors.New("บัญชีถูกระงับการใช้งาน")
	}

	if s.sessions != nil {
		_ = s.sessions.ClearLoginFailures(ctx, lockKey)
	}

	var tenantIDStr string
	if user.TenantID != nil {
		tenantIDStr = user.TenantID.String()
	}

	pair, err := security.IssueTokenPair(s.secCfg, security.IssueInput{
		UserID:   user.ID.String(),
		Role:     user.Role,
		AppScope: security.ScopeRider,
		TenantID: tenantIDStr,
	})
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	if s.sessions != nil {
		err = s.sessions.CreateRefreshSession(ctx, pair.RefreshJTI, security.RefreshRecord{
			UserID:    user.ID.String(),
			FamilyID:  pair.FamilyID,
			AppScope:  security.ScopeRider,
			Role:      user.Role,
			TenantID:  tenantIDStr,
			ExpiresAt: pair.RefreshExpiresAt,
		}, s.secCfg.RefreshTokenTTL)
		if err != nil {
			return nil, errors.New("failed to persist session")
		}
	}

	return &Response{
		Token:        pair.AccessToken,
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		Role:         user.Role,
		User:         user,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (s *service) recordFailure(ctx context.Context, key string) {
	if s.sessions == nil {
		return
	}
	_, _ = s.sessions.RecordLoginFailure(ctx, key, 5, 5*time.Minute, 5*time.Minute)
}
