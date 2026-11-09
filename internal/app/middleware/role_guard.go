package middleware

import (
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
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

// OrgAdminOrSuper อนุญาตผู้บริหารระดับมหาวิทยาลัย (org_admin) หรือ Super Admin
func OrgAdminOrSuper() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || (role != "super_admin" && role != "admin" && role != "org_admin") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Organization Admin or Super Admin access required",
			})
		}
		return c.Next()
	}
}

// TenantAdminOnly อนุญาตเฉพาะ Super Admin, Org Admin หรือ Tenant Admin ของมหาวิทยาลัย
func TenantAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || (role != "super_admin" && role != "admin" && role != "org_admin" && role != "tenant_admin") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Tenant Admin or Super Admin access required",
			})
		}
		return c.Next()
	}
}

// TenantStaffOrAdmin อนุญาตทั้ง Staff, Tenant Admin, Org Admin และ Super Admin
func TenantStaffOrAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || (role != "super_admin" && role != "admin" && role != "org_admin" && role != "tenant_admin" && role != "tenant_staff") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Staff or Admin access required",
			})
		}
		return c.Next()
	}
}

// CampusSlugGuard ป้องกันไม่ให้ Tenant Admin หรือ Staff เข้าถึงข้อมูลของวิทยาเขตอื่น (org_admin เข้าได้ทุกวิทยาเขตในสังกัด)
func CampusSlugGuard(tenantRepo domainRepo.TenantRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role == "super_admin" || role == "admin" {
			return c.Next()
		}

		slug := c.Params("slug")
		if slug == "" {
			return c.Next()
		}

		tenant, err := tenantRepo.GetBySlug(c.Context(), slug)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "campus tenant not found",
			})
		}

		userOrgID, _ := c.Locals("organization_id").(string)
		if role == "org_admin" && userOrgID != "" && tenant.OrganizationID != nil {
			if tenant.OrganizationID.String() == userOrgID {
				return c.Next()
			}
		}

		userTenantID, _ := c.Locals("tenant_id").(string)
		if userTenantID == "" || userTenantID != tenant.ID.String() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: You do not have access to this campus",
			})
		}

		return c.Next()
	}
}
