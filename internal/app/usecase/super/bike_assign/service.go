package bike_assign

import (
	"context"
	"errors"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Request struct {
	BikeID     string  `json:"bike_id"`
	TenantID   *string `json:"tenant_id"`
	TenantSlug *string `json:"tenant_slug"`
}

type Service struct {
	repo domainRepo.TenantRepository
}

func NewService(repo domainRepo.TenantRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Execute(ctx context.Context, req Request) error {
	if req.BikeID == "" {
		return errors.New("bike_id is required")
	}

	var tenantUUID *uuid.UUID
	rawTarget := req.TenantID
	if rawTarget == nil || *rawTarget == "" {
		rawTarget = req.TenantSlug
	}

	if rawTarget != nil && *rawTarget != "" && *rawTarget != "none" {
		if parsed, err := uuid.Parse(*rawTarget); err == nil {
			tenantUUID = &parsed
		} else {
			// Fallback: resolve by tenant slug (e.g. 'ku', 'cmu')
			t, err := s.repo.GetBySlug(ctx, *rawTarget)
			if err != nil {
				return errors.New("target tenant not found")
			}
			tenantUUID = &t.ID
		}
	}

	return s.repo.AssignBike(ctx, req.BikeID, tenantUUID)
}
