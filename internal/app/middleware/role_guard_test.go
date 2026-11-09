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
func (m *mockTenantRepo) List(context.Context) ([]*entity.Tenant, error) { return nil, nil }
func (m *mockTenantRepo) UpdateRates(context.Context, uuid.UUID, float64, float64, float64) error {
	return nil
}
func (m *mockTenantRepo) AssignBike(context.Context, string, *uuid.UUID) error { return nil }
func (m *mockTenantRepo) ListBikesByTenant(context.Context, uuid.UUID) ([]*entity.BikeData, error) {
	return nil, nil
}
func (m *mockTenantRepo) SwapBattery(context.Context, string) error { return nil }

func TestRoleGuardsAndCampusIsolation(t *testing.T) {
	kuID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cmuID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := &mockTenantRepo{
		tenants: map[string]*entity.Tenant{
			"ku":  {ID: kuID, Slug: "ku", Name: "KU"},
			"cmu": {ID: cmuID, Slug: "cmu", Name: "CMU"},
		},
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("role", c.Get("X-Test-Role"))
		c.Locals("tenant_id", c.Get("X-Test-Tenant"))
		return c.Next()
	})
	app.Get("/super", SuperAdminOnly(), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	app.Get("/campus/:slug/admin", CampusSlugGuard(repo), TenantAdminOnly(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/campus/:slug/ops", CampusSlugGuard(repo), TenantStaffOrAdmin(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	cases := []struct {
		name     string
		path     string
		role     string
		tenantID string
		wantCode int
	}{
		{"super admin accesses super", "/super", "super_admin", "", fiber.StatusOK},
		{"tenant admin blocked from super", "/super", "tenant_admin", kuID.String(), fiber.StatusForbidden},
		{"ku admin accesses ku admin route", "/campus/ku/admin", "tenant_admin", kuID.String(), fiber.StatusOK},
		{"ku admin blocked from cmu", "/campus/cmu/admin", "tenant_admin", kuID.String(), fiber.StatusForbidden},
		{"ku staff blocked from ku admin route", "/campus/ku/admin", "tenant_staff", kuID.String(), fiber.StatusForbidden},
		{"ku staff allowed on ku ops route", "/campus/ku/ops", "tenant_staff", kuID.String(), fiber.StatusOK},
		{"super admin allowed on any campus", "/campus/cmu/admin", "super_admin", "", fiber.StatusOK},
	}

	for _, tc := range cases {
		req := httptest.NewRequest("GET", tc.path, nil)
		req.Header.Set("X-Test-Role", tc.role)
		if tc.tenantID != "" {
			req.Header.Set("X-Test-Tenant", tc.tenantID)
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
