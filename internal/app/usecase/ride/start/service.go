package start

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/pkg/utils/geo"
	"github.com/google/uuid"
)

type Request struct {
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	BikeID     string    `json:"bike_id" validate:"required"`
	CampusSlug string    `json:"campus_slug"`
	Lat        float64   `json:"lat"`
	Lon        float64   `json:"lon"`
}

type Service interface {
	Start(ctx context.Context, req Request) (*entity.Ride, error)
}

type service struct {
	rideRepo   repository.RideRepository
	userRepo   repository.UserRepository
	bikeRepo   repository.BikeRepository
	tenantRepo repository.TenantRepository
	zoneRepo   repository.ZoneRepository
}

func NewService(
	rideRepo repository.RideRepository,
	userRepo repository.UserRepository,
	bikeRepo repository.BikeRepository,
	tenantRepo repository.TenantRepository,
	zoneRepo repository.ZoneRepository,
) Service {
	return &service{
		rideRepo:   rideRepo,
		userRepo:   userRepo,
		bikeRepo:   bikeRepo,
		tenantRepo: tenantRepo,
		zoneRepo:   zoneRepo,
	}
}

func (s *service) Start(ctx context.Context, req Request) (*entity.Ride, error) {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	isStaff := user.Role == "staff" || user.Role == "admin" ||
		user.Role == "tenant_staff" || user.Role == "tenant_admin" ||
		user.Role == "org_admin" || user.Role == "super_admin"

	if !isStaff {
		balance, _ := strconv.ParseFloat(user.WalletBalance, 64)
		if balance < 20.0 {
			return nil, errors.New("insufficient balance (minimum 20 THB required)")
		}
	}

	active, _ := s.rideRepo.GetActiveRide(ctx, req.UserID)
	if active != nil {
		return nil, errors.New("you already have an ongoing ride")
	}

	bike, err := s.bikeRepo.GetByID(ctx, req.BikeID)
	if err != nil {
		return nil, errors.New("bike not found")
	}

	if !isStaff {
		if bike.Status != "available" {
			return nil, errors.New("bike is not available")
		}
		if bike.Battery < 10 {
			return nil, errors.New("bike battery is too low")
		}
	}

	tenant, err := s.resolveOperatingTenant(ctx, bike, req)
	if err != nil {
		return nil, err
	}

	rideType := "ride"
	if isStaff {
		rideType = "service"
	}

	rideID := uuid.New()
	ride := &entity.Ride{
		ID:            rideID,
		UserID:        req.UserID,
		BikeID:        req.BikeID,
		TenantID:      tenant.ID,
		StartLat:      bike.Lat,
		StartLon:      bike.Lon,
		UnlockFee:     fmt.Sprintf("%.2f", tenant.UnlockFee),
		RatePerMinute: fmt.Sprintf("%.2f", tenant.PerMinuteRate),
		Status:        "ongoing",
		Type:          rideType,
	}

	if err := s.rideRepo.Create(ctx, ride); err != nil {
		return nil, errors.New("failed to start ride")
	}

	_ = s.bikeRepo.UpdateRideState(ctx, bike.BikeID, "in_use", &rideID)
	return ride, nil
}

// resolveOperatingTenant attributes the ride (and therefore revenue) to the campus
// where service actually occurs: operating GPS zone -> requested QR campus_slug -> bike home tenant.
func (s *service) resolveOperatingTenant(ctx context.Context, bike *entity.BikeData, req Request) (*entity.Tenant, error) {
	pt := geo.Point{Lat: req.Lat, Lon: req.Lon}
	if pt.Lat == 0 && pt.Lon == 0 {
		pt = geo.Point{Lat: bike.Lat, Lon: bike.Lon}
	}
	if (pt.Lat != 0 || pt.Lon != 0) && s.zoneRepo != nil {
		if zones, err := s.zoneRepo.ListAll(ctx); err == nil {
			for _, z := range zones {
				if z.TenantID == nil || (z.Type != "riding_limit" && z.Type != "parking") {
					continue
				}
				if pointInsideZone(pt, z.Boundary, z.Radius) {
					if t, err := s.tenantRepo.GetByID(ctx, *z.TenantID); err == nil && t.Status == "active" {
						return t, nil
					}
				}
			}
		}
	}
	if req.CampusSlug != "" {
		if t, err := s.tenantRepo.GetBySlug(ctx, req.CampusSlug); err == nil {
			return t, nil
		}
	}
	if bike.TenantID != nil {
		if t, err := s.tenantRepo.GetByID(ctx, *bike.TenantID); err == nil {
			return t, nil
		}
	}
	return nil, errors.New("bike is not assigned to any campus tenant")
}

func pointInsideZone(pt geo.Point, boundary string, radius float64) bool {
	var points []geo.Point
	if err := json.Unmarshal([]byte(boundary), &points); err != nil || len(points) == 0 {
		var single geo.Point
		if err := json.Unmarshal([]byte(boundary), &single); err != nil || (single.Lat == 0 && single.Lon == 0) {
			return false
		}
		points = []geo.Point{single}
	}
	if radius > 0 {
		return geo.IsPointInCircle(pt, points[0], radius)
	}
	return geo.IsPointInPolygon(pt, points)
}
