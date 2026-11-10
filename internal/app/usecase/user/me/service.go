package me

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Response struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone_number"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	WalletBalance string `json:"wallet_balance"`
	ImageURL      string `json:"image_url"`
}

type Service interface {
	GetMyProfile(ctx context.Context, userID string) (*Response, error)
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) GetMyProfile(ctx context.Context, userID string) (*Response, error) {
	// แปลงจาก string ID เป็น UUID
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id format")
	}

	user, err := s.userRepo.GetByID(ctx, uID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &Response{
		ID:            user.ID.String(),
		Username:      user.Username,
		Email:         user.Email,
		PhoneNumber:   user.PhoneNumber,
		Role:          user.Role,
		Status:        user.Status,
		WalletBalance: user.WalletBalance,
		ImageURL:      user.ImageURL,
	}, nil
}
