package inventory_manage

import (
	"context"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type mockTenantRepo struct {
	repository.TenantRepository
	tenant *entity.Tenant
}

func (m *mockTenantRepo) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	return m.tenant, nil
}
func (m *mockTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	return m.tenant, nil
}

type mockPartRepo struct {
	parts []*entity.SparePart
}

func (m *mockPartRepo) ListByTenant(ctx context.Context, t uuid.UUID, c string) ([]*entity.SparePart, error) {
	return m.parts, nil
}
func (m *mockPartRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.SparePart, error) {
	if len(m.parts) > 0 {
		return m.parts[0], nil
	}
	return nil, nil
}
func (m *mockPartRepo) GetByCode(ctx context.Context, t uuid.UUID, c string) (*entity.SparePart, error) {
	return nil, nil
}
func (m *mockPartRepo) Create(ctx context.Context, p *entity.SparePart) error {
	m.parts = append(m.parts, p)
	return nil
}
func (m *mockPartRepo) AdjustStock(ctx context.Context, t, p uuid.UUID, d int, ct, ref, n, a string) (*entity.SparePart, error) {
	if len(m.parts) > 0 {
		m.parts[0].Quantity += d
		return m.parts[0], nil
	}
	return nil, nil
}
func (m *mockPartRepo) ListUsageLogs(ctx context.Context, t uuid.UUID, p *uuid.UUID, l int) ([]*entity.PartUsageLog, error) {
	return []*entity.PartUsageLog{}, nil
}

func TestGetInventory_CalculatesSummary(t *testing.T) {
	tID := uuid.New()
	tenant := &entity.Tenant{ID: tID, Slug: "ku", Name: "Kasetsart"}
	parts := []*entity.SparePart{
		{ID: uuid.New(), TenantID: tID, PartName: "Battery", Quantity: 10, CostPerUnit: 1000.0, MinStock: 5},
		{ID: uuid.New(), TenantID: tID, PartName: "Brake Pad", Quantity: 2, CostPerUnit: 150.0, MinStock: 5, IsLowStock: true},
	}

	svc := NewService(&mockTenantRepo{tenant: tenant}, &mockPartRepo{parts: parts}, nil)
	summary, err := svc.GetInventory(context.Background(), "ku", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalSKUs != 2 {
		t.Fatalf("expected 2 skus, got %d", summary.TotalSKUs)
	}
	if summary.TotalUnits != 12 {
		t.Fatalf("expected 12 units, got %d", summary.TotalUnits)
	}
	if summary.LowStockCount != 1 {
		t.Fatalf("expected 1 low stock, got %d", summary.LowStockCount)
	}
	if summary.TotalValuation != 10300.0 {
		t.Fatalf("expected 10300.0 valuation, got %f", summary.TotalValuation)
	}
}
