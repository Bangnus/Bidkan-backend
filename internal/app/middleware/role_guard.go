package middleware

import (
	"slices"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
)

// SuperAdminOnly อนุญาตเฉพาะผู้ดูแลระดับแพลตฟอร์ม BidKan (super_admin) เท่านั้น
func SuperAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := GetAuthRole(c)
		if role != "super_admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Super Admin access required",
			})
		}
		return c.Next()
	}
}

// OrgAdminOnly อนุญาตเฉพาะผู้บริหารระดับมหาวิทยาลัย (org_admin)
func OrgAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := GetAuthRole(c)
		if role != "org_admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Organization Admin access required",
			})
		}
		return c.Next()
	}
}

// OrgSlugGuard ป้องกันไม่ให้เข้าถึงข้อมูลขององค์กร/มหาวิทยาลัยอื่น
func OrgSlugGuard(orgRepo domainRepo.OrganizationRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := GetAuthRole(c)
		if role != "org_admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Organization Admin access required",
			})
		}

		orgSlug := c.Params("orgSlug")
		if orgSlug == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Missing organization scope",
			})
		}

		org, err := orgRepo.GetBySlug(c.Context(), orgSlug)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "organization not found",
			})
		}

		userOrgID := GetAuthOrgID(c)
		if userOrgID == "" || userOrgID != org.ID.String() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: You do not have access to this organization",
			})
		}
		return c.Next()
	}
}

// TenantAdminOnly อนุญาตเฉพาะ Org Admin หรือ Tenant Admin ของมหาวิทยาลัย (ไม่อนุญาต Super Admin)
func TenantAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := GetAuthRole(c)
		if role != "org_admin" && role != "tenant_admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Campus or Organization Admin access required",
			})
		}
		return c.Next()
	}
}

// TenantStaffOrAdmin อนุญาต Staff, Tenant Admin และ Org Admin (ไม่อนุญาต Super Admin)
func TenantStaffOrAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := GetAuthRole(c)
		if role != "org_admin" && role != "tenant_admin" && role != "tenant_staff" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Staff or Campus Admin access required",
			})
		}
		return c.Next()
	}
}

// CampusSlugGuard บังคับขอบเขตการเข้าถึงข้อมูลวิทยาเขตตามสิทธิ์ (บล็อก super_admin เด็ดขาด)
func CampusSlugGuard(tenantRepo domainRepo.TenantRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := GetAuthRole(c)
		slug := c.Params("slug")
		if slug == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Missing campus scope",
			})
		}

		tenant, err := tenantRepo.GetBySlug(c.Context(), slug)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "campus tenant not found",
			})
		}

		userOrgID := GetAuthOrgID(c)
		if role == "org_admin" && userOrgID != "" && tenant.OrganizationID != nil {
			if tenant.OrganizationID.String() == userOrgID {
				return c.Next()
			}
		}

		userTenantID := GetAuthTenantID(c)
		if (role == "tenant_admin" || role == "tenant_staff") && userTenantID != "" && userTenantID == tenant.ID.String() {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden: You do not have access to this campus",
		})
	}
}

// BackofficeRoleGuard ตรวจสอบบทบาทที่อนุญาตสำหรับ Backoffice Endpoints
func BackofficeRoleGuard(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := GetAuthRole(c)
		if role == "" || !slices.Contains(allowedRoles, role) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: insufficient role permissions",
			})
		}
		return c.Next()
	}
}
