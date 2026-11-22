package dispute_test

import (
	"context"
	"testing"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/ride/dispute"
	"github.com/google/uuid"
)

type mockRideRepo struct {
	domainRepo.RideRepository
	disputed bool
}

func (m *mockRideRepo) DisputePenalty(ctx context.Context, rideID, userID uuid.UUID, reason string) error {
	m.disputed = true
	return nil
}

func TestDisputeRideService_Dispute(t *testing.T) {
	rRepo := &mockRideRepo{}
	svc := dispute.NewService(rRepo)

	err := svc.DisputeRide(context.Background(), uuid.New(), uuid.New(), "จอดในโซนแล้วแต่ระบบหักเงิน")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rRepo.disputed {
		t.Errorf("expected dispute to be recorded")
	}
}
