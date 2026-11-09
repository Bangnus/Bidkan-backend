package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type HardwareOrderPostgresRepository struct {
	queries *sqlc.Queries
}

func NewHardwareOrderPostgresRepository(db *sql.DB) domainRepo.HardwareOrderRepository {
	return &HardwareOrderPostgresRepository{queries: sqlc.New(db)}
}

func (r *HardwareOrderPostgresRepository) Create(ctx context.Context, o *entity.HardwareOrder) error {
	return r.queries.CreateHardwareOrder(ctx, sqlc.CreateHardwareOrderParams{
		ID:          o.ID,
		TenantID:    o.TenantID,
		OrderType:   o.OrderType,
		Quantity:    o.Quantity,
		UnitPrice:   fmt.Sprintf("%.2f", o.UnitPrice),
		TotalAmount: fmt.Sprintf("%.2f", o.TotalAmount),
		Note:        toNullString(o.Note),
		Status:      o.Status,
		CreatedBy:   toNullUUID(o.CreatedBy),
	})
}

func (r *HardwareOrderPostgresRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.HardwareOrder, error) {
	rows, err := r.queries.ListHardwareOrdersByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.HardwareOrder, 0, len(rows))
	for _, row := range rows {
		res = append(res, buildHardwareOrder(
			row.ID, row.TenantID, row.TenantSlug, row.TenantName, row.OrderType,
			row.Quantity, row.UnitPrice, row.TotalAmount, row.Note, row.Status, row.CreatedBy,
			row.CreatedAt, row.UpdatedAt,
		))
	}
	return res, nil
}

func (r *HardwareOrderPostgresRepository) ListAll(ctx context.Context) ([]*entity.HardwareOrder, error) {
	rows, err := r.queries.ListAllHardwareOrders(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.HardwareOrder, 0, len(rows))
	for _, row := range rows {
		res = append(res, buildHardwareOrder(
			row.ID, row.TenantID, row.TenantSlug, row.TenantName, row.OrderType,
			row.Quantity, row.UnitPrice, row.TotalAmount, row.Note, row.Status, row.CreatedBy,
			row.CreatedAt, row.UpdatedAt,
		))
	}
	return res, nil
}

func (r *HardwareOrderPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.queries.UpdateHardwareOrderStatus(ctx, sqlc.UpdateHardwareOrderStatusParams{
		ID:     id,
		Status: status,
	})
}

func buildHardwareOrder(
	id, tenantID uuid.UUID,
	slug, name, orderType string,
	qty int32,
	unitPriceStr, totalStr string,
	note sql.NullString,
	status string,
	createdBy uuid.NullUUID,
	createdAt, updatedAt time.Time,
) *entity.HardwareOrder {
	unitPrice, _ := strconv.ParseFloat(unitPriceStr, 64)
	totalAmount, _ := strconv.ParseFloat(totalStr, 64)
	var byPtr *uuid.UUID
	if createdBy.Valid {
		u := createdBy.UUID
		byPtr = &u
	}
	return &entity.HardwareOrder{
		ID:          id,
		TenantID:    tenantID,
		TenantSlug:  slug,
		TenantName:  name,
		OrderType:   orderType,
		Quantity:    qty,
		UnitPrice:   unitPrice,
		TotalAmount: totalAmount,
		Note:        note.String,
		Status:      status,
		CreatedBy:   byPtr,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
