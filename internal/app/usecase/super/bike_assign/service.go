package bike_assign

import (
	"context"
	"errors"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type Request struct {
	BikeID   string  `json:"bike_id"`
	TenantID *string `json:"tenant_id"` // can be nil or UUID string
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
	if req.TenantID != nil && *req.TenantID != "" && *req.TenantID != "none" {
		parsed, err := uuid.Parse(*req.TenantID)
		if err != nil {
			return errors.New("invalid tenant_id format")
		}
		tenantUUID = &parsed
	}

	return s.repo.AssignBike(ctx, req.BikeID, tenantUUID)
}

