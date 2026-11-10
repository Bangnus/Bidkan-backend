package middleware

import (
	"strings"
	"time"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/security"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthDeps struct {
	Config   *security.Config
	Sessions security.SessionStore
	UserRepo domainRepo.UserRepository
}

func AuthMiddleware(deps AuthDeps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString, ok := extractBearerToken(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid Authorization header. Use 'Bearer <token>'",
			})
		}

		claims, err := security.ParseAndVerifyToken(deps.Config, tokenString, security.TokenTypeAccess, "")
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		if deps.Sessions != nil {
			revoked, err := deps.Sessions.IsAccessBlacklisted(c.Context(), claims.ID)
			if err == nil && revoked {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token has been revoked",
				})
			}
		}

		role := claims.Role
		tenantID := claims.TenantID
		orgID := claims.OrganizationID

		if deps.UserRepo != nil {
			uid, err := uuid.Parse(claims.UserID)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user subject"})
			}
			user, err := deps.UserRepo.GetByID(c.Context(), uid)
			if err != nil || user == nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User account not found"})
			}
			if user.Status != "active" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Account is suspended or inactive"})
			}
			if user.Role != claims.Role {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Session permissions changed; please sign in again"})
			}
			role = user.Role
			if user.TenantID != nil {
				tenantID = user.TenantID.String()
			}
			if user.OrganizationID != nil {
				orgID = user.OrganizationID.String()
			}
		}

		setAuthLocals(c, claims, role, tenantID, orgID)
		return c.Next()
	}
}

func OptionalAuthMiddleware(deps AuthDeps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString, ok := extractBearerToken(c)
		if !ok {
			return c.Next()
		}
		claims, err := security.ParseAndVerifyToken(deps.Config, tokenString, security.TokenTypeAccess, "")
		if err != nil {
			return c.Next()
		}
		if deps.Sessions != nil {
			if revoked, err := deps.Sessions.IsAccessBlacklisted(c.Context(), claims.ID); err == nil && revoked {
				return c.Next()
			}
		}
		setAuthLocals(c, claims, claims.Role, claims.TenantID, claims.OrganizationID)
		return c.Next()
	}
}

func RequireScope(expectedScope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		scope := GetAuthScope(c)
		if scope != expectedScope {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: token not valid for this application scope",
			})
		}
		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return BackofficeRoleGuard("super_admin", "org_admin", "tenant_admin")
}

func extractBearerToken(c *fiber.Ctx) (string, bool) {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return "", false
	}
	return token, true
}

func setAuthLocals(c *fiber.Ctx, claims *security.AuthClaims, role, tenantID, orgID string) {
	c.Locals("user_id", claims.UserID)
	c.Locals("role", role)
	c.Locals("app_scope", claims.AppScope)
	c.Locals("jti", claims.ID)
	c.Locals("family_id", claims.FamilyID)
	if claims.ExpiresAt != nil {
		c.Locals("token_exp", claims.ExpiresAt.Time)
	} else {
		c.Locals("token_exp", time.Now())
	}
	if tenantID != "" {
		c.Locals("tenant_id", tenantID)
	}
	if orgID != "" {
		c.Locals("organization_id", orgID)
	}
}
