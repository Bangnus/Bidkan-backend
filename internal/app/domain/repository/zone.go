package repository

import (
	"context"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type ZoneRepository interface {
    Create(ctx context.Context, zone *entity.Zone) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Zone, error)
    ListAll(ctx context.Context) ([]entity.Zone, error)
    ListByType(ctx context.Context, zoneType string) ([]entity.Zone, error)
    Update(ctx context.Context, zone *entity.Zone) error
    Delete(ctx context.Context, id uuid.UUID) error
}
