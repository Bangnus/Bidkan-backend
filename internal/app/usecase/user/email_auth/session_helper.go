package email_auth

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
)

func (s *service) issueSessionResponse(ctx context.Context, user *entity.User) (*VerifyEmailResponse, error) {
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

	return &VerifyEmailResponse{
		Token:        pair.AccessToken,
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		Role:         user.Role,
		User:         user,
	}, nil
}
