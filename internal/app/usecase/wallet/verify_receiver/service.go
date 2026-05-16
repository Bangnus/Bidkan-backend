package verify_receiver

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Response struct {
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
}

type Service interface {
	Execute(ctx context.Context, phone string) (*Response, error)
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) Execute(ctx context.Context, phone string) (*Response, error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, errors.New("user not found with this phone number")
	}

	return &Response{
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
	}, nil
}
