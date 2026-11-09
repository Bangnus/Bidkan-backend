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
		TenantID: toNullUUID(zone.TenantID),
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
	return r.mapZones(zones), nil
}

func (r *zonePostgresRepository) ListByType(ctx context.Context, zoneType string) ([]entity.Zone, error) {
	zones, err := r.queries.ListZonesByType(ctx, zoneType)
	if err != nil {
		return nil, err
	}
	return r.mapZones(zones), nil
}

func (r *zonePostgresRepository) ListByTenantAndType(ctx context.Context, tenantID uuid.UUID, zoneType string) ([]entity.Zone, error) {
	zones, err := r.queries.ListZonesByTenantAndType(ctx, sqlc.ListZonesByTenantAndTypeParams{
		TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
		Type:     zoneType,
	})
	if err != nil {
		return nil, err
	}
	return r.mapZones(zones), nil
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

func (r *zonePostgresRepository) mapZones(rows []sqlc.Zone) []entity.Zone {
	res := make([]entity.Zone, 0, len(rows))
	for _, z := range rows {
		res = append(res, *r.mapZone(z))
	}
	return res
}

func (r *zonePostgresRepository) mapZone(z sqlc.Zone) *entity.Zone {
	var tID *uuid.UUID
	if z.TenantID.Valid {
		u := z.TenantID.UUID
		tID = &u
	}
	return &entity.Zone{
		ID:        z.ID,
		TenantID:  tID,
		Name:      z.Name,
		Type:      z.Type,
		Boundary:  z.Boundary,
		Radius:    z.Radius,
		CreatedAt: z.CreatedAt,
		UpdatedAt: z.UpdatedAt,
	}
}
