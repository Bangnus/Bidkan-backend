package tenant_create

import (
	"context"
	"errors"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Request struct {
	Slug               string  `json:"slug"`
	Name               string  `json:"name"`
	PlatformFeePercent float64 `json:"platform_fee_percent"`
	UnlockFee          float64 `json:"unlock_fee"`
	PerMinuteRate      float64 `json:"per_minute_rate"`
}

type Service struct {
	repo domainRepo.TenantRepository
}

func NewService(repo domainRepo.TenantRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Execute(ctx context.Context, req Request) (*entity.Tenant, error) {
	if req.Slug == "" || req.Name == "" {
		return nil, errors.New("slug and name are required")
	}
	if req.PlatformFeePercent <= 0 {
		req.PlatformFeePercent = 15.00
	}
	if req.PerMinuteRate <= 0 {
		req.PerMinuteRate = 1.00
	}

	tenant := &entity.Tenant{
		ID:                        uuid.New(),
		Slug:                      req.Slug,
		Name:                      req.Name,
		PlatformFeePercent:        req.PlatformFeePercent,
		UnlockFee:                 req.UnlockFee,
		PerMinuteRate:             req.PerMinuteRate,
		Status:                    "active",
		StripeOnboardingCompleted: false,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

