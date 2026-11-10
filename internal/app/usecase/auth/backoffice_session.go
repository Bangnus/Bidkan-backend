package auth

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) issueBackofficeSession(ctx context.Context, user *entity.User, familyID string) (*security.TokenPair, fiber.Map, error) {
	profile, meta := h.buildAuthProfile(ctx, user)
	pair, err := security.IssueTokenPair(h.secCfg, security.IssueInput{
		UserID:         user.ID.String(),
		Role:           user.Role,
		AppScope:       security.ScopeBackoffice,
		FamilyID:       familyID,
		TenantID:       meta.tenantID,
		TenantSlug:     meta.tenantSlug,
		OrganizationID: meta.orgID,
		OrgSlug:        meta.orgSlug,
		AllowedSlugs:   meta.allowedSlugs,
	})
	if err != nil {
		return nil, nil, err
	}

	if h.sessions != nil {
		err = h.sessions.CreateRefreshSession(ctx, pair.RefreshJTI, security.RefreshRecord{
			UserID:         user.ID.String(),
			FamilyID:       pair.FamilyID,
			AppScope:       security.ScopeBackoffice,
			Role:           user.Role,
			TenantID:       meta.tenantID,
			TenantSlug:     meta.tenantSlug,
			OrganizationID: meta.orgID,
			OrgSlug:        meta.orgSlug,
			AllowedSlugs:   meta.allowedSlugs,
			ExpiresAt:      pair.RefreshExpiresAt,
		}, h.secCfg.RefreshTokenTTL)
		if err != nil {
			return nil, nil, err
		}
	}

	return pair, profile, nil
}

func (h *Handler) HandleRefresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "refresh_token is required"})
	}

	claims, err := security.ParseAndVerifyToken(h.secCfg, strings.TrimSpace(req.RefreshToken), security.TokenTypeRefresh, security.ScopeBackoffice)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired refresh token"})
	}

	if h.sessions != nil {
		if _, err := h.sessions.RotateRefreshSession(c.Context(), claims.ID); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "session revoked or expired"})
		}
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token subject"})
	}
	user, err := h.userRepo.GetByID(c.Context(), uid)
	if err != nil || user == nil || user.Status != "active" || !isBackofficeRole(user.Role) {
		if h.sessions != nil && claims.FamilyID != "" {
			_ = h.sessions.RevokeFamily(c.Context(), claims.FamilyID)
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "account unavailable"})
	}

	pair, profile, err := h.issueBackofficeSession(c.Context(), user, claims.FamilyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to rotate session"})
	}

	profile["token"] = pair.AccessToken
	profile["access_token"] = pair.AccessToken
	profile["refresh_token"] = pair.RefreshToken
	profile["expires_in"] = pair.ExpiresIn
	return c.Status(fiber.StatusOK).JSON(profile)
}

func (h *Handler) HandleLogout(c *fiber.Ctx) error {
	var req LogoutRequest
	_ = c.BodyParser(&req)

	jti := middleware.GetAuthJTI(c)
	exp := middleware.GetAuthExp(c)
	familyID := middleware.GetAuthFamilyID(c)

	if h.sessions != nil {
		if jti != "" {
			ttl := time.Until(exp)
			if ttl < time.Second {
				ttl = time.Minute
			}
			_ = h.sessions.BlacklistAccessToken(c.Context(), jti, ttl)
		}
		if familyID != "" {
			_ = h.sessions.RevokeFamily(c.Context(), familyID)
		}
		if rt := strings.TrimSpace(req.RefreshToken); rt != "" {
			if claims, err := security.ParseAndVerifyToken(h.secCfg, rt, security.TokenTypeRefresh, security.ScopeBackoffice); err == nil && claims.FamilyID != "" {
				_ = h.sessions.RevokeFamily(c.Context(), claims.FamilyID)
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Logged out successfully"})
}

func setRetryHeader(c *fiber.Ctx, d time.Duration) {
	secs := int(d.Seconds()) + 1
	c.Set("Retry-After", strconv.Itoa(secs))
}

