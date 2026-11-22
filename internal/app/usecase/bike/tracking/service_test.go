package tracking

import (
	"context"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type mockBikeRepo struct {
	bike *entity.BikeData
}

func (m *mockBikeRepo) Create(ctx context.Context, b entity.BikeData) error                    { return nil }
func (m *mockBikeRepo) GetByID(ctx context.Context, id string) (*entity.BikeData, error)    { return m.bike, nil }
func (m *mockBikeRepo) ListAll(ctx context.Context) ([]entity.BikeData, error)             { return nil, nil }
func (m *mockBikeRepo) ListAvailable(ctx context.Context) ([]entity.BikeData, error)       { return nil, nil }
func (m *mockBikeRepo) UpdateStatus(ctx context.Context, data entity.BikeData) error        { return nil }
func (m *mockBikeRepo) SaveLocation(ctx context.Context, data entity.BikeData) error        { return nil }
func (m *mockBikeRepo) UpdateRideState(ctx context.Context, id, s string, r *uuid.UUID) error { return nil }

type mockAuditRepo struct {
	logged []*entity.AuditLog
}

func (m *mockAuditRepo) LogAction(ctx context.Context, log *entity.AuditLog) error {
	m.logged = append(m.logged, log)
	return nil
}
func (m *mockAuditRepo) ListAuditLogs(ctx context.Context, t uuid.UUID, a string, l int) ([]*entity.AuditLog, error) {
	return m.logged, nil
}

type mockPublisher struct {
	topics []string
}

func (m *mockPublisher) Publish(topic string, payload interface{}) error {
	m.topics = append(m.topics, topic)
	return nil
}

func TestProcessTracking_TheftAlarm(t *testing.T) {
	tenantID := uuid.New()
	bike := &entity.BikeData{
		BikeID:   "BK-001",
		Status:   "available",
		Lat:      13.750000,
		Lon:      100.500000,
		TenantID: &tenantID,
	}

	bikeRepo := &mockBikeRepo{bike: bike}
	auditRepo := &mockAuditRepo{}
	pub := &mockPublisher{}

	svc := NewService(bikeRepo, nil, nil, auditRepo, pub)

	// Bike moved 100+ meters while status is "available"
	movedData := entity.BikeData{
		BikeID:  "BK-001",
		Status:  "available",
		Lat:     13.751000,
		Lon:     100.501000,
		Battery: 80,
	}

	err := svc.ProcessTracking(context.Background(), movedData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pub.topics) == 0 {
		t.Fatal("expected MQTT beep alarm to be published")
	}
	if len(auditRepo.logged) == 0 || auditRepo.logged[0].Action != "theft_alarm_triggered" {
		t.Fatal("expected theft_alarm_triggered audit log")
	}
}

type mockZoneRepo struct {
	zones []entity.Zone
}

func (m *mockZoneRepo) Create(ctx context.Context, z *entity.Zone) error                 { return nil }
func (m *mockZoneRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Zone, error)  { return nil, nil }
func (m *mockZoneRepo) ListAll(ctx context.Context) ([]entity.Zone, error)               { return m.zones, nil }
func (m *mockZoneRepo) ListByType(ctx context.Context, t string) ([]entity.Zone, error)  { return m.zones, nil }
func (m *mockZoneRepo) ListByTenantAndType(ctx context.Context, t uuid.UUID, typ string) ([]entity.Zone, error) {
	return m.zones, nil
}
func (m *mockZoneRepo) Update(ctx context.Context, z *entity.Zone) error { return nil }
func (m *mockZoneRepo) Delete(ctx context.Context, id uuid.UUID) error  { return nil }

func TestProcessTracking_GeofenceBreachCutoff(t *testing.T) {
	tenantID := uuid.New()
	rideID := uuid.New().String()
	bike := &entity.BikeData{
		BikeID:        "BK-002",
		Status:        "in_use",
		Lat:           13.750000,
		Lon:           100.500000,
		TenantID:      &tenantID,
		CurrentRideID: &rideID,
	}

	bikeRepo := &mockBikeRepo{bike: bike}
	auditRepo := &mockAuditRepo{}
	pub := &mockPublisher{}
	zoneRepo := &mockZoneRepo{
		zones: []entity.Zone{
			{
				ID:       uuid.New(),
				TenantID: &tenantID,
				Type:     "riding_limit",
				Boundary: `{"lat":13.750000,"lon":100.500000}`,
				Radius:   100.0,
			},
		},
	}

	svc := NewService(bikeRepo, nil, zoneRepo, auditRepo, pub)

	movedData := entity.BikeData{
		BikeID:  "BK-002",
		Status:  "in_use",
		Lat:     13.755000,
		Lon:     100.505000,
		Battery: 80,
	}

	err := svc.ProcessTracking(context.Background(), movedData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pub.topics) == 0 {
		t.Fatal("expected MQTT lock/cutoff command to be published")
	}
	if len(auditRepo.logged) == 0 || auditRepo.logged[0].Action != "geofence_breach_cutoff" {
		t.Fatal("expected geofence_breach_cutoff audit log")
	}
}
