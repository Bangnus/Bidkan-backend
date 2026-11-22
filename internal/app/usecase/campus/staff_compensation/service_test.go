package staff_compensation_test

import (
	"context"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/staff_compensation"
	"github.com/google/uuid"
)

type mockTenantRepo struct {
	domainRepo.TenantRepository
	tenant *entity.Tenant
}

func (m *mockTenantRepo) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	if m.tenant != nil {
		return m.tenant, nil
	}
	return &entity.Tenant{ID: uuid.New(), Slug: slug, Name: "KU"}, nil
}

type mockStaffCompRepo struct {
	config   *entity.StaffCompensationConfig
	earnings []*entity.StaffEarningSummary
	myEarn   *entity.MyEarningsSummary
	tasks    []*entity.StaffTaskLog
}

func (m *mockStaffCompRepo) GetCompensationConfig(ctx context.Context, tenantID uuid.UUID) (*entity.StaffCompensationConfig, error) {
	if m.config != nil {
		return m.config, nil
	}
	return &entity.StaffCompensationConfig{TenantID: tenantID, PayModel: "per_task", BatterySwapFee: 20, RepairTaskFee: 50}, nil
}
func (m *mockStaffCompRepo) UpsertCompensationConfig(ctx context.Context, cfg *entity.StaffCompensationConfig) error {
	m.config = cfg
	return nil
}
func (m *mockStaffCompRepo) LogStaffTask(ctx context.Context, log *entity.StaffTaskLog) error {
	m.tasks = append(m.tasks, log)
	return nil
}
func (m *mockStaffCompRepo) ListStaffEarnings(ctx context.Context, tenantID uuid.UUID) ([]*entity.StaffEarningSummary, error) {
	return m.earnings, nil
}
func (m *mockStaffCompRepo) GetStaffEarningsByUserID(ctx context.Context, tenantID, staffUserID uuid.UUID) (*entity.MyEarningsSummary, error) {
	if m.myEarn != nil {
		return m.myEarn, nil
	}
	return &entity.MyEarningsSummary{TodayEarnings: 100, PayModel: "per_task"}, nil
}

func TestStaffCompensationService_GetAndUpdateConfig(t *testing.T) {
	tRepo := &mockTenantRepo{}
	cRepo := &mockStaffCompRepo{}
	svc := staff_compensation.NewService(tRepo, cRepo)

	cfg, err := svc.GetConfig(context.Background(), "ku")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PayModel != "per_task" {
		t.Errorf("expected per_task, got %s", cfg.PayModel)
	}

	updated, err := svc.UpdateConfig(context.Background(), "ku", staff_compensation.UpdateConfigRequest{
		PayModel:       "daily",
		DailyBaseWage:  450,
		BatterySwapFee: 25,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.PayModel != "daily" || updated.DailyBaseWage != 450 {
		t.Errorf("expected daily 450, got %s %.2f", updated.PayModel, updated.DailyBaseWage)
	}
}

func TestStaffCompensationService_Earnings(t *testing.T) {
	tRepo := &mockTenantRepo{}
	cRepo := &mockStaffCompRepo{
		earnings: []*entity.StaffEarningSummary{
			{Username: "tech01", TotalSwaps: 10, TotalEarnings: 200},
		},
	}
	svc := staff_compensation.NewService(tRepo, cRepo)

	list, err := svc.ListStaffEarnings(context.Background(), "ku")
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 item, got %d, err: %v", len(list), err)
	}
	if list[0].TotalEarnings != 200 {
		t.Errorf("expected 200, got %.2f", list[0].TotalEarnings)
	}

	my, err := svc.GetMyEarnings(context.Background(), "ku", uuid.New())
	if err != nil || my.TodayEarnings != 100 {
		t.Fatalf("expected 100 today earnings, got %.2f", my.TodayEarnings)
	}
}
