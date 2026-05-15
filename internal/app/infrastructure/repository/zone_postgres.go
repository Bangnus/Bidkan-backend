package repository

import (
	"context"
	"database/sql"

	"encoding/json"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type zonePostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewZonePostgresRepository(db *sql.DB) repository.ZoneRepository {
	return &zonePostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *zonePostgresRepository) Create(ctx context.Context, zone *entity.Zone) error {
	return r.queries.CreateZone(ctx, sqlc.CreateZoneParams{
		ID:   zone.ID,
		Name: zone.Name,
		Type: zone.Type,
		Boundary: pqtype.NullRawMessage{
			RawMessage: json.RawMessage(zone.Boundary),
			Valid:      zone.Boundary != "",
		},
		Radius: sql.NullFloat64{Float64: zone.Radius, Valid: true},
	})
}

func (r *zonePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Zone, error) {
	res, err := r.queries.GetZone(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.Zone{
		ID:        res.ID,
		Name:      res.Name,
		Type:      res.Type,
		Boundary:  string(res.Boundary.RawMessage),
		Radius:    res.Radius.Float64,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}, nil
}

func (r *zonePostgresRepository) ListAll(ctx context.Context) ([]entity.Zone, error) {
	zones, err := r.queries.ListZones(ctx)
	if err != nil {
		return nil, err
	}
	var res []entity.Zone
	for _, z := range zones {
		res = append(res, entity.Zone{
			ID:        z.ID,
			Name:      z.Name,
			Type:      z.Type,
			Boundary:  string(z.Boundary.RawMessage),
			Radius:    z.Radius.Float64,
			CreatedAt: z.CreatedAt,
			UpdatedAt: z.UpdatedAt,
		})
	}
	return res, nil
}
