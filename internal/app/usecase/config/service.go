package config

import (
	"context"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Request struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type Service interface {
	SetConfig(ctx context.Context, req Request) error
	GetConfig(ctx context.Context, key string) (string, error)
}

type service struct {
	repo repository.ConfigRepository
}

func NewService(repo repository.ConfigRepository) Service {
	return &service{repo: repo}
}

func (s *service) SetConfig(ctx context.Context, req Request) error {
	return s.repo.SetConfig(ctx, req.Key, req.Value)
}

func (s *service) GetConfig(ctx context.Context, key string) (string, error) {
	return s.repo.GetConfig(ctx, key)
}
