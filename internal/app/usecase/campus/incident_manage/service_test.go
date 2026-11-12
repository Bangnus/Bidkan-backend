package incident_manage

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

type mockIncidentRepo struct {
	incidents []*entity.IncidentReport
}

func (m *mockIncidentRepo) ListByTenant(ctx context.Context, t uuid.UUID, s, sev string) ([]*entity.IncidentReport, error) {
	return m.incidents, nil
}
func (m *mockIncidentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.IncidentReport, error) {
	if len(m.incidents) > 0 {
		return m.incidents[0], nil
	}
	return nil, nil
}
func (m *mockIncidentRepo) Create(ctx context.Context, inc *entity.IncidentReport) error {
	m.incidents = append(m.incidents, inc)
	return nil
}
func (m *mockIncidentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, st string, a float64, c, n string) (*entity.IncidentReport, error) {
	if len(m.incidents) > 0 {
		m.incidents[0].InsuranceStatus = st
		m.incidents[0].ApprovedAmount = a
		return m.incidents[0], nil
	}
	return nil, nil
}

func TestListIncidents_CalculatesSummary(t *testing.T) {
	tID := uuid.New()
	tenant := &entity.Tenant{ID: tID, Slug: "ku", Name: "Kasetsart"}
	incidents := []*entity.IncidentReport{
		{
			ID:              uuid.New(),
			TenantID:        tID,
			ReporterName:    "Rider A",
			InsuranceStatus: "filed",
			ClaimAmount:     5000.0,
			ApprovedAmount:  0.0,
		},
		{
			ID:              uuid.New(),
			TenantID:        tID,
			ReporterName:    "Rider B",
			InsuranceStatus: "approved",
			ClaimAmount:     3000.0,
			ApprovedAmount:  3000.0,
		},
	}

	svc := NewService(&mockTenantRepo{tenant: tenant}, &mockIncidentRepo{incidents: incidents}, nil)
	summary, err := svc.ListIncidents(context.Background(), "ku", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalIncidents != 2 {
		t.Fatalf("expected 2 incidents, got %d", summary.TotalIncidents)
	}
	if summary.PendingClaims != 1 {
		t.Fatalf("expected 1 pending claim, got %d", summary.PendingClaims)
	}
	if summary.TotalClaimedAmount != 8000.0 {
		t.Fatalf("expected 8000.0 claimed, got %f", summary.TotalClaimedAmount)
	}
	if summary.TotalSettledAmount != 3000.0 {
		t.Fatalf("expected 3000.0 settled, got %f", summary.TotalSettledAmount)
	}
}
