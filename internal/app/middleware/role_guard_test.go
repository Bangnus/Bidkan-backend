package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type mockTenantRepo struct {
	tenants map[string]*entity.Tenant
}

func (m *mockTenantRepo) Create(context.Context, *entity.Tenant) error { return nil }
func (m *mockTenantRepo) GetByID(context.Context, uuid.UUID) (*entity.Tenant, error) {
	return nil, errors.New("not implemented")
}
func (m *mockTenantRepo) GetBySlug(_ context.Context, slug string) (*entity.Tenant, error) {
	t, ok := m.tenants[slug]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}
func (m *mockTenantRepo) GetByStripeAccountID(context.Context, string) (*entity.Tenant, error) {
	return nil, errors.New("not implemented")
}
func (m *mockTenantRepo) UpdateStripeStatus(context.Context, uuid.UUID, string, bool) error {
	return nil
}
func (m *mockTenantRepo) List(context.Context) ([]*entity.Tenant, error) { return nil, nil }
func (m *mockTenantRepo) ListByOrganization(context.Context, uuid.UUID) ([]*entity.Tenant, error) {
	return nil, nil
}
func (m *mockTenantRepo) UpdateRates(context.Context, uuid.UUID, float64, float64, float64) error {
	return nil
}
func (m *mockTenantRepo) UpdateRatesFull(context.Context, uuid.UUID, float64, float64, float64, float64, float64, bool) error {
	return nil
}
func (m *mockTenantRepo) RecalculateTieredGP(context.Context, uuid.UUID) (float64, error) {
	return 18.0, nil
}
func (m *mockTenantRepo) CountFleetForTenant(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockTenantRepo) AssignOrganization(context.Context, uuid.UUID, *uuid.UUID) error {
	return nil
}
func (m *mockTenantRepo) AssignBike(context.Context, string, *uuid.UUID) error { return nil }
func (m *mockTenantRepo) ListBikesByTenant(context.Context, uuid.UUID) ([]*entity.BikeData, error) {
	return nil, nil
}
func (m *mockTenantRepo) SwapBattery(context.Context, string) error { return nil }

type mockOrgRepo struct {
	orgs map[string]*entity.Organization
}

func (m *mockOrgRepo) Create(context.Context, *entity.Organization) error { return nil }
func (m *mockOrgRepo) GetByID(context.Context, uuid.UUID) (*entity.Organization, error) {
	return nil, errors.New("not implemented")
}
func (m *mockOrgRepo) GetBySlug(_ context.Context, slug string) (*entity.Organization, error) {
	o, ok := m.orgs[slug]
	if !ok {
		return nil, errors.New("not found")
	}
	return o, nil
}
func (m *mockOrgRepo) List(context.Context) ([]*entity.Organization, error) { return nil, nil }
func (m *mockOrgRepo) UpdateStatus(context.Context, uuid.UUID, string) error { return nil }

func TestRoleGuardsAndStrictIsolation(t *testing.T) {
	kuOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	kuID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	kuBkkID := uuid.MustParse("11111111-1111-1111-1111-111111111112")
	cmuID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	tenantRepo := &mockTenantRepo{
		tenants: map[string]*entity.Tenant{
			"ku":     {ID: kuID, OrganizationID: &kuOrgID, Slug: "ku", Name: "KU"},
			"ku-bkk": {ID: kuBkkID, OrganizationID: &kuOrgID, Slug: "ku-bkk", Name: "KU BKK"},
			"cmu":    {ID: cmuID, Slug: "cmu", Name: "CMU"},
		},
	}
	orgRepo := &mockOrgRepo{
		orgs: map[string]*entity.Organization{
			"ku-org": {ID: kuOrgID, Slug: "ku-org", Name: "KU"},
		},
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("role", c.Get("X-Test-Role"))
		c.Locals("app_scope", c.Get("X-Test-Scope"))
		c.Locals("tenant_id", c.Get("X-Test-Tenant"))
		c.Locals("organization_id", c.Get("X-Test-Org"))
		return c.Next()
	})
	app.Get("/super", RequireScope("backoffice"), SuperAdminOnly(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/org/:orgSlug/overview", RequireScope("backoffice"), OrgAdminOnly(), OrgSlugGuard(orgRepo), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/campus/:slug/admin", RequireScope("backoffice"), CampusSlugGuard(tenantRepo), TenantAdminOnly(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/campus/:slug/ops", RequireScope("backoffice"), CampusSlugGuard(tenantRepo), TenantStaffOrAdmin(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	cases := []struct {
		name     string
		path     string
		role     string
		scope    string
		tenantID string
		orgID    string
		wantCode int
	}{
		{"super admin accesses super", "/super", "super_admin", "backoffice", "", "", fiber.StatusOK},
		{"rider token blocked from super", "/super", "super_admin", "rider", "", "", fiber.StatusForbidden},
		{"super admin blocked from campus admin", "/campus/cmu/admin", "super_admin", "backoffice", "", "", fiber.StatusForbidden},
		{"super admin blocked from campus ops", "/campus/ku/ops", "super_admin", "backoffice", "", "", fiber.StatusForbidden},
		{"super admin blocked from org overview", "/org/ku-org/overview", "super_admin", "backoffice", "", "", fiber.StatusForbidden},
		{"tenant admin blocked from super", "/super", "tenant_admin", "backoffice", kuID.String(), "", fiber.StatusForbidden},
		{"ku admin accesses ku admin route", "/campus/ku/admin", "tenant_admin", "backoffice", kuID.String(), "", fiber.StatusOK},
		{"ku admin blocked from cmu", "/campus/cmu/admin", "tenant_admin", "backoffice", kuID.String(), "", fiber.StatusForbidden},
		{"ku org admin accesses own org", "/org/ku-org/overview", "org_admin", "backoffice", kuID.String(), kuOrgID.String(), fiber.StatusOK},
		{"ku org admin accesses ku-bkk", "/campus/ku-bkk/admin", "org_admin", "backoffice", kuID.String(), kuOrgID.String(), fiber.StatusOK},
		{"ku org admin blocked from cmu", "/campus/cmu/admin", "org_admin", "backoffice", kuID.String(), kuOrgID.String(), fiber.StatusForbidden},
		{"ku staff blocked from ku admin route", "/campus/ku/admin", "tenant_staff", "backoffice", kuID.String(), "", fiber.StatusForbidden},
		{"ku staff allowed on ku ops route", "/campus/ku/ops", "tenant_staff", "backoffice", kuID.String(), "", fiber.StatusOK},
	}

	for _, tc := range cases {
		req := httptest.NewRequest("GET", tc.path, nil)
		req.Header.Set("X-Test-Role", tc.role)
		req.Header.Set("X-Test-Scope", tc.scope)
		if tc.tenantID != "" {
			req.Header.Set("X-Test-Tenant", tc.tenantID)
		}
		if tc.orgID != "" {
			req.Header.Set("X-Test-Org", tc.orgID)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if resp.StatusCode != tc.wantCode {
			t.Errorf("%s: got status %d, want %d", tc.name, resp.StatusCode, tc.wantCode)
		}
	}
}
