package inventory_manage

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type AdjustStockReq struct {
	PartID      uuid.UUID `json:"part_id" validate:"required"`
	Delta       int       `json:"delta" validate:"required"` // positive = stock in, negative = stock out
	ChangeType  string    `json:"change_type" validate:"required"` // in, out, maintenance, damaged
	ReferenceID string    `json:"reference_id"`
	Notes       string    `json:"notes"`
	ActorName   string    `json:"actor_name"`
}

type CreatePartReq struct {
	PartName    string  `json:"part_name" validate:"required"`
	PartCode    string  `json:"part_code" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	Quantity    int     `json:"quantity"`
	Unit        string  `json:"unit"`
	MinStock    int     `json:"min_stock"`
	CostPerUnit float64 `json:"cost_per_unit"`
}

type Service struct {
	tenantRepo repository.TenantRepository
	partRepo   repository.SparePartRepository
	auditRepo  repository.AuditRepository
}

func NewService(
	tenantRepo repository.TenantRepository,
	partRepo repository.SparePartRepository,
	auditRepo repository.AuditRepository,
) *Service {
	return &Service{
		tenantRepo: tenantRepo,
		partRepo:   partRepo,
		auditRepo:  auditRepo,
	}
}

func (s *Service) GetInventory(ctx context.Context, slug, category string) (*entity.InventorySummary, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus not found")
	}

	parts, err := s.partRepo.ListByTenant(ctx, tenant.ID, category)
	if err != nil {
		return nil, err
	}

	summary := &entity.InventorySummary{
		TotalSKUs: len(parts),
		Parts:     parts,
	}

	for _, p := range parts {
		summary.TotalUnits += p.Quantity
		summary.TotalValuation += float64(p.Quantity) * p.CostPerUnit
		if p.IsLowStock {
			summary.LowStockCount++
		}
	}
	return summary, nil
}

func (s *Service) AdjustStock(ctx context.Context, slug string, req AdjustStockReq) (*entity.SparePart, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus not found")
	}

	if req.Delta == 0 {
		return nil, errors.New("delta cannot be zero")
	}
	if req.ActorName == "" {
		req.ActorName = "Campus Staff"
	}

	part, err := s.partRepo.AdjustStock(ctx, tenant.ID, req.PartID, req.Delta, req.ChangeType, req.ReferenceID, req.Notes, req.ActorName)
	if err != nil {
		return nil, err
	}

	if s.auditRepo != nil {
		_ = s.auditRepo.LogAction(ctx, &entity.AuditLog{
			TenantID:   tenant.ID,
			ActorName:  req.ActorName,
			ActorRole:  "tenant_staff",
			Action:     "inventory_adjusted",
			TargetType: "spare_part",
			TargetID:   part.ID.String(),
			Details:    fmt.Sprintf("Stock adjusted by %d units (%s). Balance after: %d. Ref: %s", req.Delta, req.ChangeType, part.Quantity, req.ReferenceID),
			IPAddress:  "127.0.0.1",
		})
	}

	return part, nil
}

func (s *Service) CreatePart(ctx context.Context, slug string, req CreatePartReq) (*entity.SparePart, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus not found")
	}

	if req.Unit == "" {
		req.Unit = "ชิ้น"
	}
	if req.MinStock <= 0 {
		req.MinStock = 5
	}

	part := &entity.SparePart{
		ID:          uuid.New(),
		TenantID:    tenant.ID,
		PartName:    req.PartName,
		PartCode:    req.PartCode,
		Category:    req.Category,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		MinStock:    req.MinStock,
		CostPerUnit: req.CostPerUnit,
	}

	if err := s.partRepo.Create(ctx, part); err != nil {
		return nil, err
	}
	return s.partRepo.GetByID(ctx, part.ID)
}

func (s *Service) ListUsageLogs(ctx context.Context, slug string, partID *uuid.UUID, limit int) ([]*entity.PartUsageLog, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus not found")
	}
	return s.partRepo.ListUsageLogs(ctx, tenant.ID, partID, limit)
}
