package create

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/pkg/utils/geo"
	"github.com/google/uuid"
)

type Request struct {
	Name     string      `json:"name" validate:"required"`
	Type     string      `json:"type" validate:"required"` // "P" หรือ "Area"
	Boundary []geo.Point `json:"boundary"`                 // ขอบเขต Polygon
	Radius   float64     `json:"radius"`                   // ถ้าระบุ Radius จะใช้เป็นวงกลม (ใช้พิกัดแรกใน Boundary เป็นจุดศูนย์กลาง)
}

type Service interface {
	Create(ctx context.Context, req Request) error
}

type service struct {
	zoneRepo repository.ZoneRepository
}

func NewService(zoneRepo repository.ZoneRepository) Service {
	return &service{zoneRepo: zoneRepo}
}

func (s *service) Create(ctx context.Context, req Request) error {
	if req.Type != "P" && req.Type != "Area" {
		return errors.New("invalid zone type. must be 'P' or 'Area'")
	}

	if len(req.Boundary) == 0 {
		return errors.New("boundary cannot be empty")
	}

	boundaryJSON, err := json.Marshal(req.Boundary)
	if err != nil {
		return errors.New("failed to parse boundary")
	}

	zone := entity.Zone{
		ID:       uuid.New(),
		Name:     req.Name,
		Type:     req.Type,
		Boundary: string(boundaryJSON),
		Radius:   req.Radius,
	}

	return s.zoneRepo.Create(ctx, &zone)
}
