package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetAuthUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	if raw, ok := c.Locals("user_id").(string); ok && raw != "" {
		uid, err := uuid.Parse(raw)
		if err == nil {
			return uid, true
		}
	}
	if uid, ok := c.Locals("user_id").(uuid.UUID); ok && uid != uuid.Nil {
		return uid, true
	}
	return uuid.Nil, false
}

func GetAuthRole(c *fiber.Ctx) string {
	role, _ := c.Locals("role").(string)
	return role
}

func GetAuthScope(c *fiber.Ctx) string {
	scope, _ := c.Locals("app_scope").(string)
	return scope
}

func GetAuthTenantID(c *fiber.Ctx) string {
	tid, _ := c.Locals("tenant_id").(string)
	return tid
}

func GetAuthOrgID(c *fiber.Ctx) string {
	oid, _ := c.Locals("organization_id").(string)
	return oid
}

func GetAuthJTI(c *fiber.Ctx) string {
	jti, _ := c.Locals("jti").(string)
	return jti
}

func GetAuthFamilyID(c *fiber.Ctx) string {
	fid, _ := c.Locals("family_id").(string)
	return fid
}

func GetAuthExp(c *fiber.Ctx) time.Time {
	exp, _ := c.Locals("token_exp").(time.Time)
	return exp
}

