package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
)

type zonePostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewZoneRepository(db *sql.DB) repository.ZoneRepository {
	return &zonePostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *zonePostgresRepository) Create(ctx context.Context, zone *entity.Zone) error {
	return r.queries.CreateZone(ctx, sqlc.CreateZoneParams{
		ID:       zone.ID,
		Name:     zone.Name,
		Type:     zone.Type,
		Boundary: zone.Boundary,
		Radius:   zone.Radius,
	})
}

func (r *zonePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Zone, error) {
	z, err := r.queries.GetZone(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.mapZone(z), nil
}

func (r *zonePostgresRepository) ListAll(ctx context.Context) ([]entity.Zone, error) {
	zones, err := r.queries.ListZones(ctx)
	if err != nil {
		return nil, err
	}
	var res []entity.Zone
	for _, z := range zones {
		res = append(res, *r.mapZone(z))
	}
	return res, nil
}

func (r *zonePostgresRepository) ListByType(ctx context.Context, zoneType string) ([]entity.Zone, error) {
	zones, err := r.queries.ListZonesByType(ctx, zoneType)
	if err != nil {
		return nil, err
	}
	var res []entity.Zone
	for _, z := range zones {
		res = append(res, *r.mapZone(z))
	}
	return res, nil
}

func (r *zonePostgresRepository) Update(ctx context.Context, zone *entity.Zone) error {
	return r.queries.UpdateZone(ctx, sqlc.UpdateZoneParams{
		ID:       zone.ID,
		Name:     zone.Name,
		Type:     zone.Type,
		Boundary: zone.Boundary,
		Radius:   zone.Radius,
	})
}

func (r *zonePostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteZone(ctx, id)
}

func (r *zonePostgresRepository) mapZone(z sqlc.Zone) *entity.Zone {
	return &entity.Zone{
		ID:        z.ID,
		Name:      z.Name,
		Type:      z.Type,
		Boundary:  z.Boundary,
		Radius:    z.Radius,
		CreatedAt: z.CreatedAt,
		UpdatedAt: z.UpdatedAt,
	}
}
