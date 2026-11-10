package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	ScopeBackoffice = "backoffice"
	ScopeRider      = "rider"

	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	TokenIssuer = "bidkan-api"
)

type AuthClaims struct {
	UserID         string   `json:"user_id"`
	Role           string   `json:"role"`
	AppScope       string   `json:"app_scope"`
	TokenType      string   `json:"token_type"`
	FamilyID       string   `json:"family_id,omitempty"`
	TenantID       string   `json:"tenant_id,omitempty"`
	TenantSlug     string   `json:"tenant_slug,omitempty"`
	OrganizationID string   `json:"organization_id,omitempty"`
	OrgSlug        string   `json:"org_slug,omitempty"`
	AllowedSlugs   []string `json:"allowed_slugs,omitempty"`
	jwt.RegisteredClaims
}

type IssueInput struct {
	UserID         string
	Role           string
	AppScope       string
	FamilyID       string
	TenantID       string
	TenantSlug     string
	OrganizationID string
	OrgSlug        string
	AllowedSlugs   []string
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessJTI        string
	RefreshJTI       string
	FamilyID         string
	ExpiresIn        int64
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

func IssueTokenPair(cfg *Config, in IssueInput) (*TokenPair, error) {
	if cfg == nil || len(cfg.JWTSecret) < MinSecretLength {
		return nil, errors.New("invalid security configuration")
	}
	if in.UserID == "" || in.Role == "" || in.AppScope == "" {
		return nil, errors.New("missing required claim attributes")
	}

	now := time.Now().UTC()
	familyID := in.FamilyID
	if familyID == "" {
		familyID = uuid.NewString()
	}

	accessJTI := uuid.NewString()
	refreshJTI := uuid.NewString()
	accessExp := now.Add(cfg.AccessTokenTTL)
	refreshExp := now.Add(cfg.RefreshTokenTTL)

	accessClaims := buildClaims(in, TokenTypeAccess, accessJTI, familyID, now, accessExp)
	refreshClaims := buildClaims(in, TokenTypeRefresh, refreshJTI, familyID, now, refreshExp)

	accessStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(cfg.JWTSecret)
	if err != nil {
		return nil, err
	}
	refreshStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessStr,
		RefreshToken:     refreshStr,
		AccessJTI:        accessJTI,
		RefreshJTI:       refreshJTI,
		FamilyID:         familyID,
		ExpiresIn:        int64(cfg.AccessTokenTTL.Seconds()),
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func buildClaims(in IssueInput, tokenType, jti, familyID string, now, exp time.Time) AuthClaims {
	return AuthClaims{
		UserID:         in.UserID,
		Role:           in.Role,
		AppScope:       in.AppScope,
		TokenType:      tokenType,
		FamilyID:       familyID,
		TenantID:       in.TenantID,
		TenantSlug:     in.TenantSlug,
		OrganizationID: in.OrganizationID,
		OrgSlug:        in.OrgSlug,
		AllowedSlugs:   in.AllowedSlugs,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    TokenIssuer,
			Subject:   in.UserID,
			Audience:  jwt.ClaimStrings{in.AppScope},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
}

func ParseAndVerifyToken(cfg *Config, tokenStr, expectedType, expectedScope string) (*AuthClaims, error) {
	if cfg == nil || len(cfg.JWTSecret) < MinSecretLength {
		return nil, errors.New("invalid security configuration")
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(TokenIssuer),
		jwt.WithExpirationRequired(),
	)

	claims := &AuthClaims{}
	token, err := parser.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (interface{}, error) {
		return cfg.JWTSecret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	if claims.ID == "" || claims.UserID == "" || claims.Role == "" {
		return nil, errors.New("incomplete token claims")
	}
	if expectedType != "" && claims.TokenType != expectedType {
		return nil, errors.New("unexpected token type")
	}
	if expectedScope != "" && claims.AppScope != expectedScope {
		return nil, errors.New("invalid token application scope")
	}
	return claims, nil
}

