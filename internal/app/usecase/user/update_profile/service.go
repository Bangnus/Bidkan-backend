package update_profile

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Request struct {
	UserID   uuid.UUID `json:"-"` // ดึงจาก Token
	ImageURL string    `json:"image_url" validate:"required,url"`
}

type Service interface {
	UpdateProfileImage(ctx context.Context, req Request) error
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) UpdateProfileImage(ctx context.Context, req Request) error {
	return s.userRepo.UpdateProfileImage(ctx, req.UserID, req.ImageURL)
}
