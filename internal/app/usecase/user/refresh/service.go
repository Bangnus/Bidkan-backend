package refresh

import (
	"context"
	"errors"
	"strings"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/google/uuid"
)

type Request struct {
	RefreshToken string `json:"refresh_token"`
}

type Response struct {
	Token        string `json:"token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Role         string `json:"role"`
}

type Service interface {
	Refresh(ctx context.Context, req Request) (*Response, error)
}

type service struct {
	userRepo repository.UserRepository
	secCfg   *security.Config
	sessions security.SessionStore
}

func NewService(userRepo repository.UserRepository, secCfg *security.Config, sessions security.SessionStore) Service {
	return &service{userRepo: userRepo, secCfg: secCfg, sessions: sessions}
}

func (s *service) Refresh(ctx context.Context, req Request) (*Response, error) {
	rt := strings.TrimSpace(req.RefreshToken)
	if rt == "" {
		return nil, errors.New("refresh_token is required")
	}

	claims, err := security.ParseAndVerifyToken(s.secCfg, rt, security.TokenTypeRefresh, security.ScopeRider)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	if s.sessions != nil {
		if _, err := s.sessions.RotateRefreshSession(ctx, claims.ID); err != nil {
			return nil, errors.New("session revoked or expired")
		}
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid token subject")
	}
	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil || user == nil || user.Status != "active" || user.Role != "user" {
		if s.sessions != nil && claims.FamilyID != "" {
			_ = s.sessions.RevokeFamily(ctx, claims.FamilyID)
		}
		return nil, errors.New("account unavailable")
	}

	var tenantIDStr string
	if user.TenantID != nil {
		tenantIDStr = user.TenantID.String()
	}

	pair, err := security.IssueTokenPair(s.secCfg, security.IssueInput{
		UserID:   user.ID.String(),
		Role:     user.Role,
		AppScope: security.ScopeRider,
		FamilyID: claims.FamilyID,
		TenantID: tenantIDStr,
	})
	if err != nil {
		return nil, errors.New("failed to rotate session")
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
			return nil, errors.New("failed to persist rotated session")
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

