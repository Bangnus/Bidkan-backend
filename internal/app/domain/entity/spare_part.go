package entity

import (
	"time"

	"github.com/google/uuid"
)

type SparePart struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	PartName    string    `json:"part_name"`
	PartCode    string    `json:"part_code"`
	Category    string    `json:"category"`
	Quantity    int       `json:"quantity"`
	Unit        string    `json:"unit"`
	MinStock    int       `json:"min_stock"`
	CostPerUnit float64   `json:"cost_per_unit"`
	IsLowStock  bool      `json:"is_low_stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PartUsageLog struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	PartID       uuid.UUID `json:"part_id"`
	PartName     string    `json:"part_name,omitempty"`
	PartCode     string    `json:"part_code,omitempty"`
	ChangeType   string    `json:"change_type"` // in, out, maintenance, damaged
	Quantity     int       `json:"quantity"`
	BalanceAfter int       `json:"balance_after"`
	ReferenceID  string    `json:"reference_id"`
	Notes        string    `json:"notes"`
	ActorName    string    `json:"actor_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type InventorySummary struct {
	TotalSKUs     int          `json:"total_skus"`
	TotalUnits    int          `json:"total_units"`
	LowStockCount int          `json:"low_stock_count"`
	TotalValuation float64     `json:"total_valuation"`
	Parts         []*SparePart `json:"parts"`
}
