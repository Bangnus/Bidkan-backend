package rebalancing_test

import (
	"context"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/rebalancing"
	"github.com/google/uuid"
)

type mockTenantRepo struct {
	domainRepo.TenantRepository
	bikes []*entity.BikeData
}

func (m *mockTenantRepo) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	return &entity.Tenant{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Slug: slug, Name: "KU"}, nil
}

func (m *mockTenantRepo) ListBikesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.BikeData, error) {
	return m.bikes, nil
}

type mockCampusOpsRepo struct {
	domainRepo.CampusOpsRepository
	zones []*entity.TenantZone
}

func (m *mockCampusOpsRepo) ListZonesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantZone, error) {
	return m.zones, nil
}

type mockRebalRepo struct {
	domainRepo.RebalancingRepository
}

func (m *mockRebalRepo) GetHeatmapPoints(ctx context.Context, tenantID uuid.UUID, limit int) ([]*entity.HeatmapPoint, error) {
	return []*entity.HeatmapPoint{
		{Lat: 13.847, Lon: 100.569, Intensity: 0.9, Type: "pickup"},
	}, nil
}

func TestRebalancingService_Overview(t *testing.T) {
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tRepo := &mockTenantRepo{
		bikes: []*entity.BikeData{
			// 12 bikes in Zone 1 (Overflow: target 8)
			{BikeID: "B1", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B2", Lat: 13.8471, Lon: 100.5691},
			{BikeID: "B3", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B4", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B5", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B6", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B7", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B8", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B9", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B10", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B11", Lat: 13.8470, Lon: 100.5690},
			{BikeID: "B12", Lat: 13.8470, Lon: 100.5690},
			// 1 bike in Zone 2 (Deficit: target 8)
			{BikeID: "B13", Lat: 13.8550, Lon: 100.5750},
		},
	}
	cRepo := &mockCampusOpsRepo{
		zones: []*entity.TenantZone{
			{ID: uuid.New(), TenantID: tenantID, Name: "Dorm 14", Type: "parking", Boundary: `[{"lat":13.8470,"lon":100.5690}]`, Radius: 50},
			{ID: uuid.New(), TenantID: tenantID, Name: "Main Gate", Type: "parking", Boundary: `[{"lat":13.8550,"lon":100.5750}]`, Radius: 50},
		},
	}
	rRepo := &mockRebalRepo{}

	svc := rebalancing.NewService(tRepo, cRepo, rRepo)
	res, err := svc.GetRebalanceOverview(context.Background(), "ku")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Zones) != 2 {
		t.Fatalf("expected 2 zones, got %d", len(res.Zones))
	}
	if res.OverflowZonesCount != 1 {
		t.Errorf("expected 1 overflow zone, got %d", res.OverflowZonesCount)
	}
	if res.DeficitZonesCount != 1 {
		t.Errorf("expected 1 deficit zone, got %d", res.DeficitZonesCount)
	}
	if len(res.DispatchTasks) == 0 {
		t.Errorf("expected dispatch tasks to be generated")
	}
	if len(res.HeatmapPoints) != 1 {
		t.Errorf("expected 1 heatmap point, got %d", len(res.HeatmapPoints))
	}
}

