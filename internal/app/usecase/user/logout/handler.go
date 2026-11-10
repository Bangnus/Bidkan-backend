package logout

import (
	"strings"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/Bangnus/Bidkan-backend/internal/app/middleware"
	"github.com/gofiber/fiber/v2"
)

type Request struct {
	RefreshToken string `json:"refresh_token"`
}

type Handler interface {
	Handle(c *fiber.Ctx) error
}

type handler struct {
	secCfg   *security.Config
	sessions security.SessionStore
}

func NewHandler(secCfg *security.Config, sessions security.SessionStore) Handler {
	return &handler{secCfg: secCfg, sessions: sessions}
}

// @Summary Logout User
// @Description Revoke the current access and refresh JWT session in Redis.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /v1/users/logout [post]
func (h *handler) Handle(c *fiber.Ctx) error {
	var req Request
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
		if rt := strings.TrimSpace(req.RefreshToken); rt != "" && h.secCfg != nil {
			if claims, err := security.ParseAndVerifyToken(h.secCfg, rt, security.TokenTypeRefresh, security.ScopeRider); err == nil && claims.FamilyID != "" {
				_ = h.sessions.RevokeFamily(c.Context(), claims.FamilyID)
			}
		}
	}

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
