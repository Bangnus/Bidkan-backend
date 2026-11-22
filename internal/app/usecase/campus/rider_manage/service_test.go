package rider_manage_test

import (
	"context"
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/campus/rider_manage"
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
	return &entity.Tenant{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Slug: slug, Name: "KU"}, nil
}

type mockRiderRepo struct {
	riders []*entity.CampusRider
	disputes []*entity.DisputedRide
	refunded bool
	verified bool
	status string
}

func (m *mockRiderRepo) ListCampusRiders(ctx context.Context, tenantID uuid.UUID, search, filter string) ([]*entity.CampusRider, error) {
	return m.riders, nil
}
func (m *mockRiderRepo) UpdateStudentVerification(ctx context.Context, userID uuid.UUID, verified bool) error {
	m.verified = verified
	return nil
}
func (m *mockRiderRepo) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	m.status = status
	return nil
}
func (m *mockRiderRepo) ListDisputedRides(ctx context.Context, tenantID uuid.UUID, status string) ([]*entity.DisputedRide, error) {
	return m.disputes, nil
}
func (m *mockRiderRepo) RefundPenalty(ctx context.Context, tenantID, rideID, userID uuid.UUID, refundAmount float64, reason string) error {
	m.refunded = true
	return nil
}

func TestRiderManageService_ListAndVerify(t *testing.T) {
	tRepo := &mockTenantRepo{}
	rRepo := &mockRiderRepo{
		riders: []*entity.CampusRider{
			{Username: "testuser", IsStudent: true, StudentVerified: false},
		},
	}
	svc := rider_manage.NewService(tRepo, rRepo)

	list, err := svc.ListRiders(context.Background(), "ku", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 rider, got %d", len(list))
	}

	uID := uuid.New()
	err = svc.VerifyStudent(context.Background(), "ku", uID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rRepo.verified {
		t.Errorf("expected verified to be true")
	}
}

func TestRiderManageService_ProcessRefund(t *testing.T) {
	tRepo := &mockTenantRepo{}
	rRepo := &mockRiderRepo{}
	svc := rider_manage.NewService(tRepo, rRepo)

	err := svc.ProcessRefund(context.Background(), "ku", rider_manage.RefundPenaltyRequest{
		RideID:       uuid.New(),
		UserID:       uuid.New(),
		RefundAmount: 50.0,
		Reason:       "จอดในช่องถูกต้องแต่สัญญาณ GPS คลาดเคลื่อน",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rRepo.refunded {
		t.Errorf("expected penalty to be refunded")
	}
}

