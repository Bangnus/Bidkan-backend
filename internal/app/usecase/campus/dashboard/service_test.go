package dashboard

import (
	"context"
	"errors"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type stubTenantRepo struct {
	tenant     *entity.Tenant
	bikes      []*entity.BikeData
	fleetCount int
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
func (s *stubTenantRepo) GetByStripeAccountID(context.Context, string) (*entity.Tenant, error) {
	return nil, errors.New("not found")
}
func (s *stubTenantRepo) UpdateStripeStatus(context.Context, uuid.UUID, string, bool) error {
	return nil
}
func (s *stubTenantRepo) List(context.Context) ([]*entity.Tenant, error) { return nil, nil }
func (s *stubTenantRepo) ListByOrganization(context.Context, uuid.UUID) ([]*entity.Tenant, error) {
	return nil, nil
}
func (s *stubTenantRepo) UpdateRates(context.Context, uuid.UUID, float64, float64, float64) error {
	return nil
}
func (s *stubTenantRepo) UpdateRatesFull(context.Context, uuid.UUID, float64, float64, float64, float64, float64, bool) error {
	return nil
}
func (s *stubTenantRepo) RecalculateTieredGP(context.Context, uuid.UUID) (float64, error) {
	return 15.0, nil
}
func (s *stubTenantRepo) CountFleetForTenant(context.Context, uuid.UUID) (int, error) {
	return s.fleetCount, nil
}
func (s *stubTenantRepo) AssignOrganization(context.Context, uuid.UUID, *uuid.UUID) error {
	return nil
}
func (s *stubTenantRepo) AssignBike(context.Context, string, *uuid.UUID) error { return nil }
func (s *stubTenantRepo) ListBikesByTenant(context.Context, uuid.UUID) ([]*entity.BikeData, error) {
	return s.bikes, nil
}
func (s *stubTenantRepo) SwapBattery(context.Context, string) error { return nil }

type stubCampusRepo struct {
	domainRepo.CampusOpsRepository
	today *entity.RideStatsSummary
}

func (c *stubCampusRepo) GetTodayStats(context.Context, uuid.UUID) (*entity.RideStatsSummary, error) {
	return c.today, nil
}

func TestDashboardServiceMetricsAndTieredGP(t *testing.T) {
	repo := &stubTenantRepo{
		tenant: &entity.Tenant{
			ID:                 uuid.New(),
			Slug:               "ku",
			Name:               "KU",
			PlatformFeePercent: 15.0,
		},
		fleetCount: 42,
		bikes: []*entity.BikeData{
			{BikeID: "BK-001", Status: "available", Battery: 90},
			{BikeID: "BK-002", Status: "in_use", Battery: 65},
			{BikeID: "BK-003", Status: "available", Battery: 12},
			{BikeID: "BK-004", Status: "maintenance", Battery: 19},
		},
	}
	campusRepo := &stubCampusRepo{
		today: &entity.RideStatsSummary{TotalRides: 12, GrossAmount: 300, GPFeeAmount: 45, NetAmount: 255},
	}

	svc := NewService(repo, campusRepo)
	res, err := svc.Execute(context.Background(), "ku")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalBikes != 4 || res.AvailableBikes != 2 || res.InUseBikes != 1 || res.LowBatteryBikes != 2 {
		t.Fatalf("unexpected fleet counters: %+v", res)
	}
	if res.TotalTripsToday != 12 || res.GrossVolumeToday != 300 || res.NetEarningsToday != 255 {
		t.Fatalf("expected real DB today stats, got: %+v", res)
	}
	if res.GPTier.CurrentGPPercent != 15.0 || res.GPTier.NextTierTarget != 101 || res.GPTier.BikesToNextTier != 59 {
		t.Fatalf("unexpected Tiered GP info: %+v", res.GPTier)
	}
}
