package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/google/uuid"
)

type SparePartPostgres struct {
	db *sql.DB
}

func NewSparePartPostgres(db *sql.DB) repository.SparePartRepository {
	return &SparePartPostgres{db: db}
}

func (r *SparePartPostgres) ListByTenant(ctx context.Context, tenantID uuid.UUID, category string) ([]*entity.SparePart, error) {
	query := `
		SELECT id, tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit, created_at, updated_at
		FROM campus_spare_parts
		WHERE tenant_id = $1`
	args := []interface{}{tenantID}

	if category != "" && category != "all" {
		query += " AND category = $2"
		args = append(args, category)
	}
	query += " ORDER BY category ASC, part_name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.SparePart
	for rows.Next() {
		var p entity.SparePart
		if err := rows.Scan(&p.ID, &p.TenantID, &p.PartName, &p.PartCode, &p.Category, &p.Quantity, &p.Unit, &p.MinStock, &p.CostPerUnit, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.IsLowStock = p.Quantity <= p.MinStock
		list = append(list, &p)
	}
	return list, nil
}

func (r *SparePartPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.SparePart, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit, created_at, updated_at
		FROM campus_spare_parts WHERE id = $1`, id)
	var p entity.SparePart
	if err := row.Scan(&p.ID, &p.TenantID, &p.PartName, &p.PartCode, &p.Category, &p.Quantity, &p.Unit, &p.MinStock, &p.CostPerUnit, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.IsLowStock = p.Quantity <= p.MinStock
	return &p, nil
}

func (r *SparePartPostgres) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*entity.SparePart, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit, created_at, updated_at
		FROM campus_spare_parts WHERE tenant_id = $1 AND part_code = $2`, tenantID, code)
	var p entity.SparePart
	if err := row.Scan(&p.ID, &p.TenantID, &p.PartName, &p.PartCode, &p.Category, &p.Quantity, &p.Unit, &p.MinStock, &p.CostPerUnit, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.IsLowStock = p.Quantity <= p.MinStock
	return &p, nil
}

func (r *SparePartPostgres) Create(ctx context.Context, part *entity.SparePart) error {
	if part.ID == uuid.Nil {
		part.ID = uuid.New()
	}
	query := `
		INSERT INTO campus_spare_parts (id, tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query, part.ID, part.TenantID, part.PartName, part.PartCode, part.Category, part.Quantity, part.Unit, part.MinStock, part.CostPerUnit)
	return err
}

func (r *SparePartPostgres) AdjustStock(ctx context.Context, tenantID, partID uuid.UUID, delta int, changeType, refID, notes, actor string) (*entity.SparePart, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentQty int
	err = tx.QueryRowContext(ctx, `
		SELECT quantity FROM campus_spare_parts
		WHERE id = $1 AND tenant_id = $2 FOR UPDATE`, partID, tenantID).Scan(&currentQty)
	if err != nil {
		return nil, fmt.Errorf("part not found: %w", err)
	}

	newQty := currentQty + delta
	if newQty < 0 {
		return nil, fmt.Errorf("insufficient stock (current: %d, requested reduction: %d)", currentQty, -delta)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE campus_spare_parts
		SET quantity = $1, updated_at = NOW()
		WHERE id = $2`, newQty, partID)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO parts_usage_logs (tenant_id, part_id, change_type, quantity, balance_after, reference_id, notes, actor_name, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		tenantID, partID, changeType, delta, newQty, refID, notes, actor)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, partID)
}

func (r *SparePartPostgres) ListUsageLogs(ctx context.Context, tenantID uuid.UUID, partID *uuid.UUID, limit int) ([]*entity.PartUsageLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `
		SELECT l.id, l.tenant_id, l.part_id, COALESCE(p.part_name, '-'), COALESCE(p.part_code, '-'),
		       l.change_type, l.quantity, l.balance_after, l.reference_id, l.notes, l.actor_name, l.created_at
		FROM parts_usage_logs l
		LEFT JOIN campus_spare_parts p ON l.part_id = p.id
		WHERE l.tenant_id = $1`
	args := []interface{}{tenantID}

	if partID != nil {
		query += " AND l.part_id = $2"
		args = append(args, *partID)
	}
	query += fmt.Sprintf(" ORDER BY l.created_at DESC LIMIT %d", limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entity.PartUsageLog
	for rows.Next() {
		var l entity.PartUsageLog
		var rID, notes sql.NullString
		if err := rows.Scan(&l.ID, &l.TenantID, &l.PartID, &l.PartName, &l.PartCode,
			&l.ChangeType, &l.Quantity, &l.BalanceAfter, &rID, &notes, &l.ActorName, &l.CreatedAt); err != nil {
			return nil, err
		}
		if rID.Valid && rID.String != "" {
			l.ReferenceID = rID.String
		} else {
			l.ReferenceID = "-"
		}
		if notes.Valid && notes.String != "" {
			l.Notes = notes.String
		} else {
			l.Notes = "-"
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
