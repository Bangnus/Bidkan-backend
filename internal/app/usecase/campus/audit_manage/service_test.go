package audit_manage_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/audit_manage"
	"github.com/google/uuid"
)

type mockTenantRepo struct {
	domainRepo.TenantRepository
}

func (m *mockTenantRepo) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	return &entity.Tenant{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Slug: slug, Name: "KU"}, nil
}

type mockCampusOpsRepo struct {
	domainRepo.CampusOpsRepository
}

func (m *mockCampusOpsRepo) ListRidesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.CampusRideDetail, error) {
	return []*entity.CampusRideDetail{
		{
			ID:              uuid.New(),
			BikeID:          "BK-01",
			Username:        "somchai",
			PhoneNumber:     "0812345678",
			DistanceKm:      2.5,
			TotalFare:       35.0,
			PlatformFee:     7.0,
			PartnerEarnings: 28.0,
			Status:          "completed",
			StartTime:       time.Now(),
		},
	}, nil
}

func (m *mockCampusOpsRepo) GetCompletedStats(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error) {
	return &entity.RideStatsSummary{
		GrossAmount: 1000.0,
		GPFeeAmount: 200.0,
		NetAmount:   800.0,
		TotalRides:  50,
	}, nil
}

func (m *mockCampusOpsRepo) GetUnsettledStats(ctx context.Context, tenantID uuid.UUID) (*entity.RideStatsSummary, error) {
	return &entity.RideStatsSummary{
		NetAmount: 150.0,
	}, nil
}

type mockStaffRepo struct {
	domainRepo.StaffCompensationRepository
}

func (m *mockStaffRepo) ListStaffEarnings(ctx context.Context, tenantID uuid.UUID) ([]*entity.StaffEarningSummary, error) {
	return []*entity.StaffEarningSummary{
		{
			StaffUserID:   uuid.New(),
			Username:      "staff1",
			PhoneNumber:   "0899999999",
			TotalSwaps:    10,
			TotalRepairs:  2,
			TotalEarnings: 300.0,
		},
	}, nil
}

type mockAuditRepo struct {
	domainRepo.AuditRepository
	logs []*entity.AuditLog
}

func (m *mockAuditRepo) ListAuditLogs(ctx context.Context, tenantID uuid.UUID, action string, limit int) ([]*entity.AuditLog, error) {
	return m.logs, nil
}
func (m *mockAuditRepo) LogAction(ctx context.Context, log *entity.AuditLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func TestAuditManageService_Exports(t *testing.T) {
	tRepo := &mockTenantRepo{}
	cRepo := &mockCampusOpsRepo{}
	sRepo := &mockStaffRepo{}
	aRepo := &mockAuditRepo{
		logs: []*entity.AuditLog{
			{ActorName: "Admin", Action: "refund_penalty", TargetType: "ride", TargetID: "R1"},
		},
	}

	svc := audit_manage.NewService(tRepo, cRepo, sRepo, aRepo)

	// Test List Audit Logs
	logs, err := svc.ListAuditLogs(context.Background(), "ku", "")
	if err != nil || len(logs) != 1 {
		t.Fatalf("failed to list audit logs: %v", err)
	}

	// Test Export Rides CSV
	ridesCSV, err := svc.ExportRidesCSV(context.Background(), "ku")
	if err != nil {
		t.Fatalf("failed to export rides csv: %v", err)
	}
	if !strings.Contains(string(ridesCSV), "BK-01") {
		t.Errorf("expected rides csv to contain BK-01")
	}

	// Test Export Revenue CSV
	revCSV, err := svc.ExportRevenueCSV(context.Background(), "ku")
	if err != nil {
		t.Fatalf("failed to export revenue csv: %v", err)
	}
	if !strings.Contains(string(revCSV), "1000.00") {
		t.Errorf("expected revenue csv to contain 1000.00")
	}

	// Test Export Staff Payouts CSV
	staffCSV, err := svc.ExportStaffPayoutsCSV(context.Background(), "ku")
	if err != nil {
		t.Fatalf("failed to export staff payouts csv: %v", err)
	}
	if !strings.Contains(string(staffCSV), "staff1") {
		t.Errorf("expected staff csv to contain staff1")
	}
}
