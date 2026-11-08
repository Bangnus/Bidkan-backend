package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SuperAdminOnly อนุญาตเฉพาะผู้ดูแลระดับแพลตฟอร์ม BidKan
func SuperAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || (role != "super_admin" && role != "admin") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Super Admin access required",
			})
		}
		return c.Next()
	}
}

// TenantAdminOnly อนุญาตเฉพาะ Super Admin หรือ Tenant Admin ของมหาวิทยาลัย
func TenantAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || (role != "super_admin" && role != "admin" && role != "tenant_admin") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Tenant Admin or Super Admin access required",
			})
		}
		return c.Next()
	}
}

// TenantStaffOrAdmin อนุญาตทั้ง Staff, Tenant Admin และ Super Admin
func TenantStaffOrAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || (role != "super_admin" && role != "admin" && role != "tenant_admin" && role != "tenant_staff") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Staff or Admin access required",
			})
		}
		return c.Next()
	}
}

// TenantScopeGuard ตรวจสอบให้มั่นใจว่า User เข้าถึงเฉพาะ Tenant ของตนเอง
func TenantScopeGuard() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		// Super Admin สามารถเข้าถึงได้ทุก Tenant
		if role == "super_admin" || role == "admin" {
			return c.Next()
		}

		userTenantID, _ := c.Locals("tenant_id").(string)
		targetTenantID := c.Params("tenantId")

		if targetTenantID != "" && userTenantID != targetTenantID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Cannot access other tenant resources",
			})
		}

		return c.Next()
	}
}

