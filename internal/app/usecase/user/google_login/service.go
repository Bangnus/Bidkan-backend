package google_login

import (
	"context"
	"errors"
	"strings"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

var (
	ErrInvalidGoogleToken = errors.New("invalid or unverified google token")
	ErrAccountRestricted  = errors.New("account is restricted or cannot sign in as rider")
)

type Request struct {
	Credential string `json:"credential" validate:"required"`
	CampusSlug string `json:"campus_slug"`
}

type Response struct {
	Token        string       `json:"token"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	Role         string       `json:"role"`
	User         *entity.User `json:"user"`
}

type GoogleClaims struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

type Service interface {
	Login(ctx context.Context, req Request) (*Response, error)
}

type service struct {
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	secCfg     *security.Config
	sessions   security.SessionStore
}

func NewService(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	secCfg *security.Config,
	sessions security.SessionStore,
) Service {
	return &service{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		secCfg:     secCfg,
		sessions:   sessions,
	}
}

func (s *service) Login(ctx context.Context, req Request) (*Response, error) {
	if strings.TrimSpace(req.Credential) == "" {
		return nil, ErrInvalidGoogleToken
	}

	claims, err := s.verifyGoogleToken(ctx, req.Credential)
	if err != nil {
		return nil, err
	}

	tenantID := s.resolveTenantID(ctx, claims.Email, req.CampusSlug)
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

	if user.Role != "user" || user.Status != "active" {
		return nil, ErrAccountRestricted
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

func (s *service) verifyGoogleToken(ctx context.Context, credential string) (*GoogleClaims, error) {
	if s.secCfg == nil || strings.TrimSpace(s.secCfg.GoogleClientID) == "" {
		return nil, errors.New("google OAuth is not configured on server")
	}

	payload, err := idtoken.Validate(ctx, credential, s.secCfg.GoogleClientID)
	if err != nil || payload == nil {
		return nil, ErrInvalidGoogleToken
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	if payload.Subject == "" || email == "" || !emailVerified {
		return nil, ErrInvalidGoogleToken
	}

	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	return &GoogleClaims{
		Sub:           payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
		Picture:       picture,
	}, nil
}

func (s *service) resolveTenantID(ctx context.Context, email, campusSlug string) *uuid.UUID {
	lowerEmail := strings.ToLower(strings.TrimSpace(email))
	slug := strings.ToLower(strings.TrimSpace(campusSlug))

	switch {
	case strings.HasSuffix(lowerEmail, "ku.th") || strings.HasSuffix(lowerEmail, "ku.ac.th"):
		slug = "ku"
	case strings.HasSuffix(lowerEmail, "cmu.ac.th"):
		slug = "cmu"
	case slug == "":
		slug = "ku"
	}

	if s.tenantRepo != nil {
		if t, err := s.tenantRepo.GetBySlug(ctx, slug); err == nil && t != nil {
			id := t.ID
			return &id
		}
		if fallback, err := s.tenantRepo.GetBySlug(ctx, "ku"); err == nil && fallback != nil {
			id := fallback.ID
			return &id
		}
	}
	return nil
}
