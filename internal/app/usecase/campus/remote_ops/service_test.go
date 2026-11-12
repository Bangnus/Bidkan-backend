package remote_ops_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/remote_ops"
	"github.com/google/uuid"
)

type mockTenantRepo struct {
	domainRepo.TenantRepository
	tenant *entity.Tenant
	bikes  []*entity.BikeData
}

func (m *mockTenantRepo) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	if m.tenant != nil {
		return m.tenant, nil
	}
	return &entity.Tenant{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Slug: slug, Name: "KU"}, nil
}

func (m *mockTenantRepo) ListBikesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.BikeData, error) {
	return m.bikes, nil
}

type mockBikeRepo struct {
	domainRepo.BikeRepository
	bike *entity.BikeData
}

func (m *mockBikeRepo) GetByID(ctx context.Context, bikeID string) (*entity.BikeData, error) {
	return m.bike, nil
}

func (m *mockBikeRepo) UpdateRideState(ctx context.Context, bikeID, status string, rideID *uuid.UUID) error {
	if m.bike != nil {
		m.bike.Status = status
	}
	return nil
}

type mockRideRepo struct {
	domainRepo.RideRepository
	ended bool
}

func (m *mockRideRepo) EndRide(ctx context.Context, id uuid.UUID, endLat, endLon, distance float64, fare, platformFee, partnerEarnings string) error {
	m.ended = true
	return nil
}

type mockMqttPub struct {
	topics   []string
	payloads []interface{}
}

func (m *mockMqttPub) Publish(topic string, payload interface{}) error {
	m.topics = append(m.topics, topic)
	m.payloads = append(m.payloads, payload)
	return nil
}

func TestRemoteOpsService_GetLiveFleet(t *testing.T) {
	now := time.Now()
	tRepo := &mockTenantRepo{
		bikes: []*entity.BikeData{
			{BikeID: "BK-01", Status: "available", Battery: 85, Lat: 13.84, Lon: 100.57, LastHeartbeat: now},
			{BikeID: "BK-02", Status: "in_use", Battery: 15, Lat: 13.85, Lon: 100.58, LastHeartbeat: now},
			{BikeID: "BK-03", Status: "maintenance", Battery: 50, Lat: 13.84, Lon: 100.57, LastHeartbeat: now.Add(-1 * time.Hour)},
		},
	}
	bRepo := &mockBikeRepo{}
	rRepo := &mockRideRepo{}
	mqttPub := &mockMqttPub{}

	svc := remote_ops.NewService(tRepo, bRepo, rRepo, mqttPub)
	res, err := svc.GetLiveFleet(context.Background(), "ku")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalBikes != 3 {
		t.Errorf("expected 3 bikes, got %d", res.TotalBikes)
	}
	if res.AvailableCount != 1 {
		t.Errorf("expected 1 available, got %d", res.AvailableCount)
	}
	if res.InUseCount != 1 {
		t.Errorf("expected 1 in use, got %d", res.InUseCount)
	}
	if res.LowBatteryCount != 1 {
		t.Errorf("expected 1 low battery, got %d", res.LowBatteryCount)
	}
	if res.OfflineCount != 1 {
		t.Errorf("expected 1 offline, got %d", res.OfflineCount)
	}
}

func TestRemoteOpsService_SendCommand(t *testing.T) {
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tRepo := &mockTenantRepo{}
	bRepo := &mockBikeRepo{
		bike: &entity.BikeData{
			BikeID:   "BK-01",
			TenantID: &tenantID,
			Status:   "available",
		},
	}
	rRepo := &mockRideRepo{}
	mqttPub := &mockMqttPub{}

	svc := remote_ops.NewService(tRepo, bRepo, rRepo, mqttPub)

	err := svc.SendCommand(context.Background(), "ku", "BK-01", "beep")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mqttPub.topics) != 1 || mqttPub.topics[0] != "bidkan/bikes/BK-01/commands" {
		t.Errorf("expected topic bidkan/bikes/BK-01/commands, got %v", mqttPub.topics)
	}
}

func TestRemoteOpsService_ForceEndRide(t *testing.T) {
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	rideIDStr := "11111111-1111-1111-1111-111111111111"
	tRepo := &mockTenantRepo{}
	bRepo := &mockBikeRepo{
		bike: &entity.BikeData{
			BikeID:        "BK-01",
			TenantID:      &tenantID,
			Status:        "in_use",
			CurrentRideID: &rideIDStr,
			Lat:           13.84,
			Lon:           100.57,
		},
	}
	rRepo := &mockRideRepo{}
	mqttPub := &mockMqttPub{}

	svc := remote_ops.NewService(tRepo, bRepo, rRepo, mqttPub)

	err := svc.ForceEndRide(context.Background(), "ku", "BK-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rRepo.ended {
		t.Errorf("expected ride to be ended")
	}
	if bRepo.bike.Status != "available" {
		t.Errorf("expected bike status available, got %s", bRepo.bike.Status)
	}
}
