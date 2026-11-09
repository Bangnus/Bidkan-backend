package dashboard

import (
	"context"
	"errors"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type stubTenantRepo struct {
	tenant *entity.Tenant
	bikes  []*entity.BikeData
}

func (s *stubTenantRepo) Create(context.Context, *entity.Tenant) error { return nil }
func (s *stubTenantRepo) GetByID(context.Context, uuid.UUID) (*entity.Tenant, error) {
	return s.tenant, nil
}
func (s *stubTenantRepo) GetBySlug(_ context.Context, slug string) (*entity.Tenant, error) {
	if s.tenant == nil || s.tenant.Slug != slug {
		return nil, errors.New("not found")
	}
	return s.tenant, nil
}
func (s *stubTenantRepo) List(context.Context) ([]*entity.Tenant, error) { return nil, nil }
func (s *stubTenantRepo) UpdateRates(context.Context, uuid.UUID, float64, float64, float64) error {
	return nil
}
func (s *stubTenantRepo) AssignBike(context.Context, string, *uuid.UUID) error { return nil }
func (s *stubTenantRepo) ListBikesByTenant(context.Context, uuid.UUID) ([]*entity.BikeData, error) {
	return s.bikes, nil
}
func (s *stubTenantRepo) SwapBattery(context.Context, string) error { return nil }

func TestDashboardServiceMetricsAndGPShare(t *testing.T) {
	repo := &stubTenantRepo{
		tenant: &entity.Tenant{
			ID:                 uuid.New(),
			Slug:               "ku",
			Name:               "KU",
			PlatformFeePercent: 15.0,
		},
		bikes: []*entity.BikeData{
			{BikeID: "BK-001", Status: "available", Battery: 90},
			{BikeID: "BK-002", Status: "in_use", Battery: 65},
			{BikeID: "BK-003", Status: "available", Battery: 12},
			{BikeID: "BK-004", Status: "maintenance", Battery: 19},
		},
	}

	svc := NewService(repo)
	res, err := svc.Execute(context.Background(), "ku")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalBikes != 4 || res.AvailableBikes != 2 || res.InUseBikes != 1 || res.LowBatteryBikes != 2 {
		t.Fatalf("unexpected fleet counters: %+v", res)
	}
	// tripsToday = 4*5 = 20, gross = 500, 85% net = 425
	if res.GrossVolumeToday != 500 || res.NetEarningsToday != 425 {
		t.Fatalf("unexpected revenue math: gross=%.2f net=%.2f", res.GrossVolumeToday, res.NetEarningsToday)
	}
}
