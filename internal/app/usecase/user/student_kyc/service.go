package student_kyc

import (
	"context"
	"errors"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Request struct {
	StudentID string `json:"student_id"`
	CardURL   string `json:"card_url"`
}

type Service struct {
	userRepo domainRepo.UserRepository
}

func NewService(userRepo domainRepo.UserRepository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) SubmitKYC(ctx context.Context, userID uuid.UUID, req Request) error {
	if req.StudentID == "" {
		return errors.New("student_id is required")
	}
	if req.CardURL == "" {
		return errors.New("card_url is required")
	}
	return s.userRepo.SubmitStudentKYC(ctx, userID, req.StudentID, req.CardURL)
}
