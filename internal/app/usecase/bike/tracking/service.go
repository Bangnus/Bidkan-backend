package tracking

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/pkg/utils/geo"
	"github.com/google/uuid"
)

type Publisher interface {
	Publish(topic string, payload interface{}) error
}

type Service interface {
	ProcessTracking(ctx context.Context, data entity.BikeData) error
}

type service struct {
	bikeRepo  repository.BikeRepository
	bikeCache repository.BikeCacheRepository
	zoneRepo  repository.ZoneRepository
	auditRepo repository.AuditRepository
	mqttPub   Publisher
}

func NewService(
	bikeRepo repository.BikeRepository,
	bikeCache repository.BikeCacheRepository,
	zoneRepo repository.ZoneRepository,
	auditRepo repository.AuditRepository,
	mqttPub Publisher,
) Service {
	return &service{
		bikeRepo:  bikeRepo,
		bikeCache: bikeCache,
		zoneRepo:  zoneRepo,
		auditRepo: auditRepo,
		mqttPub:   mqttPub,
	}
}

func (s *service) ProcessTracking(ctx context.Context, data entity.BikeData) error {
	prevBike, _ := s.bikeRepo.GetByID(ctx, data.BikeID)

	_ = s.bikeRepo.UpdateStatus(ctx, data)
	_ = s.bikeRepo.SaveLocation(ctx, data)
	if s.bikeCache != nil {
		_ = s.bikeCache.SetLatestLocation(ctx, data)
	}

	if prevBike != nil {
		s.checkTheftAlarm(ctx, prevBike, data)
		s.checkGeofenceBreach(ctx, prevBike, data)
	}
	return nil
}

func (s *service) checkTheftAlarm(ctx context.Context, prev *entity.BikeData, curr entity.BikeData) {
	if prev.Status != "available" || prev.Lat == 0 || prev.Lon == 0 || curr.Lat == 0 || curr.Lon == 0 {
		return
	}
	dist := geo.CalculateDistance(geo.Point{Lat: prev.Lat, Lon: prev.Lon}, geo.Point{Lat: curr.Lat, Lon: curr.Lon})
	if dist >= 30.0 {
		if s.mqttPub != nil {
			topic := fmt.Sprintf("bidkan/bikes/%s/commands", curr.BikeID)
			_ = s.mqttPub.Publish(topic, map[string]interface{}{
				"command":   "beep",
				"alarm":     true,
				"reason":    "unauthorized_movement",
				"bike_id":   curr.BikeID,
				"timestamp": time.Now().Unix(),
			})
		}
		if s.auditRepo != nil {
			tenantID := uuid.Nil
			if prev.TenantID != nil {
				tenantID = *prev.TenantID
			}
			_ = s.auditRepo.LogAction(ctx, &entity.AuditLog{
				TenantID:   tenantID,
				ActorName:  "SYSTEM_THEFT_DAEMON",
				ActorRole:  "system",
				Action:     "theft_alarm_triggered",
				TargetType: "bike",
				TargetID:   curr.BikeID,
				Details:    fmt.Sprintf("Unauthorized movement: moved %.1fm while available. Buzzer sounded.", dist),
				IPAddress:  "127.0.0.1",
				CreatedAt:  time.Now(),
			})
		}
	}
}

func (s *service) checkGeofenceBreach(ctx context.Context, prev *entity.BikeData, curr entity.BikeData) {
	if (curr.Status != "in_use" && prev.CurrentRideID == nil) || prev.TenantID == nil || s.zoneRepo == nil {
		return
	}
	zones, err := s.zoneRepo.ListByTenantAndType(ctx, *prev.TenantID, "riding_limit")
	if err != nil || len(zones) == 0 {
		return
	}

	pt := geo.Point{Lat: curr.Lat, Lon: curr.Lon}
	inside := false
	for _, z := range zones {
		if isInsideZone(pt, z.Boundary, z.Radius) {
			inside = true
			break
		}
	}

	if !inside {
		if s.mqttPub != nil {
			topic := fmt.Sprintf("bidkan/bikes/%s/commands", curr.BikeID)
			_ = s.mqttPub.Publish(topic, map[string]interface{}{
				"command":   "lock",
				"reason":    "out_of_boundary",
				"bike_id":   curr.BikeID,
				"timestamp": time.Now().Unix(),
			})
		}
		if s.auditRepo != nil {
			_ = s.auditRepo.LogAction(ctx, &entity.AuditLog{
				TenantID:   *prev.TenantID,
				ActorName:  "SYSTEM_GEOFENCE_DAEMON",
				ActorRole:  "system",
				Action:     "geofence_breach_cutoff",
				TargetType: "bike",
				TargetID:   curr.BikeID,
				Details:    fmt.Sprintf("Bike %s crossed outside riding limit zone. Relay cut off.", curr.BikeID),
				IPAddress:  "127.0.0.1",
				CreatedAt:  time.Now(),
			})
		}
	}
}

func isInsideZone(pt geo.Point, boundary string, radius float64) bool {
	var points []geo.Point
	if err := json.Unmarshal([]byte(boundary), &points); err != nil || len(points) == 0 {
		var single geo.Point
		if err := json.Unmarshal([]byte(boundary), &single); err != nil || (single.Lat == 0 && single.Lon == 0) {
			return false
		}
		points = []geo.Point{single}
	}
	if len(points) >= 3 {
		return geo.IsPointInPolygon(pt, points)
	}
	if radius > 0 {
		return geo.IsPointInCircle(pt, points[0], radius)
	}
	return geo.IsPointInPolygon(pt, points)
}
