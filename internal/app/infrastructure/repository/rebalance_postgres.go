package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type RebalancePostgres struct {
	db *sql.DB
}

func NewRebalancePostgres(db *sql.DB) *RebalancePostgres {
	return &RebalancePostgres{db: db}
}

func (r *RebalancePostgres) GetHeatmapPoints(ctx context.Context, tenantID uuid.UUID, limit int) ([]*entity.HeatmapPoint, error) {
	if limit <= 0 {
		limit = 300
	}
	query := `
		(SELECT start_lat as lat, start_lon as lon, 'pickup' as type
		 FROM rides
		 WHERE tenant_id = $1 AND start_lat != 0
		 ORDER BY created_at DESC LIMIT $2)
		UNION ALL
		(SELECT end_lat as lat, end_lon as lon, 'dropoff' as type
		 FROM rides
		 WHERE tenant_id = $1 AND end_lat IS NOT NULL AND end_lat != 0
		 ORDER BY created_at DESC LIMIT $2)
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit/2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*entity.HeatmapPoint
	for rows.Next() {
		p := &entity.HeatmapPoint{Intensity: 0.8}
		if err := rows.Scan(&p.Lat, &p.Lon, &p.Type); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

