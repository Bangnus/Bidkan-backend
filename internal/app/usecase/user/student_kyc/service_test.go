package student_kyc_test

import (
	"context"
	"testing"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/user/student_kyc"
	"github.com/google/uuid"
)

type mockUserRepo struct {
	domainRepo.UserRepository
	submitted bool
}

func (m *mockUserRepo) SubmitStudentKYC(ctx context.Context, id uuid.UUID, studentID, cardURL string) error {
	m.submitted = true
	return nil
}

func TestStudentKYCService_Submit(t *testing.T) {
	uRepo := &mockUserRepo{}
	svc := student_kyc.NewService(uRepo)

	err := svc.SubmitKYC(context.Background(), uuid.New(), student_kyc.Request{
		StudentID: "6510500123",
		CardURL:   "https://example.com/card.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !uRepo.submitted {
		t.Errorf("expected kyc to be submitted")
	}
}
