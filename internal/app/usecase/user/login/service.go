package login

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrTooManyAttempts = errors.New("account temporarily locked due to failed login attempts")
	ErrForbiddenScope  = errors.New("admin accounts must sign in via backoffice portal")
)

type Request struct {
	Username string `json:"username" validate:"required" example:"somchai_jaidee"`
	Password string `json:"password" validate:"required" example:"password123"`
}

type Response struct {
	Token        string `json:"token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Role         string `json:"role"`
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
	username := strings.TrimSpace(req.Username)
	lockKey := "rider:" + strings.ToLower(username)
	if s.sessions != nil {
		if locked, _, _ := s.sessions.IsLoginLocked(ctx, lockKey); locked {
			return nil, ErrTooManyAttempts
		}
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.recordFailure(ctx, lockKey)
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.recordFailure(ctx, lockKey)
		return nil, errors.New("invalid username or password")
	}

	if user.Role != "user" {
		return nil, ErrForbiddenScope
	}

	if user.Status != "active" {
		return nil, errors.New("account is not activated or is suspended")
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
	}, nil
}

func (s *service) recordFailure(ctx context.Context, key string) {
	if s.sessions == nil {
		return
	}
	_, _ = s.sessions.RecordLoginFailure(ctx, key, 5, 5*time.Minute, 5*time.Minute)
}
