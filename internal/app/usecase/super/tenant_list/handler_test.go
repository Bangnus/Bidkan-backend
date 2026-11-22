package tenant_list

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type mockTenantRepo struct {
	repository.TenantRepository
	tenants []*entity.Tenant
	total   int
}

func (m *mockTenantRepo) List(ctx context.Context) ([]*entity.Tenant, error) {
	return m.tenants, nil
}

func (m *mockTenantRepo) ListPaged(ctx context.Context, limit, offset int) ([]*entity.Tenant, int, error) {
	end := offset + limit
	if end > len(m.tenants) {
		end = len(m.tenants)
	}
	if offset > len(m.tenants) {
		return []*entity.Tenant{}, m.total, nil
	}
	return m.tenants[offset:end], m.total, nil
}

func TestTenantListHandler(t *testing.T) {
	tenants := make([]*entity.Tenant, 25)
	for i := 0; i < 25; i++ {
		tenants[i] = &entity.Tenant{
			ID:   uuid.New(),
			Name: "Campus",
			Slug: "campus",
		}
	}
	repo := &mockTenantRepo{tenants: tenants, total: 25}
	service := NewService(repo)
	handler := NewHandler(service)

	app := fiber.New()
	app.Get("/tenants", handler.Handle)

	// Case 1: Unpaged request
	req1 := httptest.NewRequest("GET", "/tenants", nil)
	resp1, err := app.Test(req1)
	if err != nil || resp1.StatusCode != fiber.StatusOK {
		t.Fatalf("unpaged request failed, status=%d, err=%v", resp1.StatusCode, err)
	}
	bodyBytes1, _ := io.ReadAll(resp1.Body)
	var out1 struct {
		Data []*entity.Tenant `json:"data"`
		Meta *entity.PageMeta `json:"meta"`
	}
	if err := json.Unmarshal(bodyBytes1, &out1); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(out1.Data) != 25 || out1.Meta != nil {
		t.Fatalf("expected 25 items and nil meta, got %d items and %+v", len(out1.Data), out1.Meta)
	}

	// Case 2: Paged request with limit
	req2 := httptest.NewRequest("GET", "/tenants?page=1&limit=10", nil)
	resp2, err := app.Test(req2)
	if err != nil || resp2.StatusCode != fiber.StatusOK {
		t.Fatalf("paged request failed, status=%d, err=%v", resp2.StatusCode, err)
	}
	bodyBytes2, _ := io.ReadAll(resp2.Body)
	var out2 struct {
		Data []*entity.Tenant `json:"data"`
		Meta *entity.PageMeta `json:"meta"`
	}
	if err := json.Unmarshal(bodyBytes2, &out2); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(out2.Data) != 10 || out2.Meta == nil || out2.Meta.Total != 25 || out2.Meta.TotalPages != 3 {
		t.Fatalf("expected 10 items, total 25, pages 3, got %d items and %+v", len(out2.Data), out2.Meta)
	}

	// Case 3: Paged request with page_size
	req3 := httptest.NewRequest("GET", "/tenants?page=2&page_size=10", nil)
	resp3, err := app.Test(req3)
	if err != nil || resp3.StatusCode != fiber.StatusOK {
		t.Fatalf("paged page_size request failed, status=%d, err=%v", resp3.StatusCode, err)
	}
	bodyBytes3, _ := io.ReadAll(resp3.Body)
	var out3 struct {
		Data []*entity.Tenant `json:"data"`
		Meta *entity.PageMeta `json:"meta"`
	}
	if err := json.Unmarshal(bodyBytes3, &out3); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(out3.Data) != 10 || out3.Meta == nil || out3.Meta.Page != 2 || out3.Meta.Total != 25 {
		t.Fatalf("expected page 2 with 10 items, got %d items and %+v", len(out3.Data), out3.Meta)
	}
}
