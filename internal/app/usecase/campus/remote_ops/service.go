package remote_ops

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/google/uuid"
)

type LiveBikeItem struct {
	BikeID        string  `json:"bike_id"`
	HardwareID    string  `json:"hardware_id"`
	Status        string  `json:"status"` // available, in_use, maintenance, offline
	Battery       int     `json:"battery"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	LastHeartbeat string  `json:"last_heartbeat"`
	IsOnline      bool    `json:"is_online"`
	CurrentRideID *string `json:"current_ride_id,omitempty"`
}

type LiveFleetResponse struct {
	TotalBikes      int             `json:"total_bikes"`
	AvailableCount  int             `json:"available_count"`
	InUseCount      int             `json:"in_use_count"`
	LowBatteryCount int             `json:"low_battery_count"`
	OfflineCount    int             `json:"offline_count"`
	Bikes           []*LiveBikeItem `json:"bikes"`
}

type Service struct {
	tenantRepo domainRepo.TenantRepository
	bikeRepo   domainRepo.BikeRepository
	rideRepo   domainRepo.RideRepository
	mqttPub    mqtt.Publisher
}

func NewService(
	tenantRepo domainRepo.TenantRepository,
	bikeRepo domainRepo.BikeRepository,
	rideRepo domainRepo.RideRepository,
	mqttPub mqtt.Publisher,
) *Service {
	return &Service{
		tenantRepo: tenantRepo,
		bikeRepo:   bikeRepo,
		rideRepo:   rideRepo,
		mqttPub:    mqttPub,
	}
}

func (s *Service) GetLiveFleet(ctx context.Context, slug string) (*LiveFleetResponse, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	bikes, err := s.tenantRepo.ListBikesByTenant(ctx, tenant.ID)
	if err != nil {
		return nil, err
	}

	res := &LiveFleetResponse{
		TotalBikes: len(bikes),
		Bikes:      make([]*LiveBikeItem, 0, len(bikes)),
	}

	now := time.Now()
	for _, b := range bikes {
		isOnline := false
		if !b.LastHeartbeat.IsZero() && now.Sub(b.LastHeartbeat) < 10*time.Minute {
			isOnline = true
		}

		status := b.Status
		if !isOnline && status != "in_use" {
			res.OfflineCount++
		}
		if b.Battery < 20 {
			res.LowBatteryCount++
		}
		if status == "available" {
			res.AvailableCount++
		} else if status == "in_use" {
			res.InUseCount++
		}

		res.Bikes = append(res.Bikes, &LiveBikeItem{
			BikeID:        b.BikeID,
			HardwareID:    b.HardwareID,
			Status:        b.Status,
			Battery:       b.Battery,
			Lat:           b.Lat,
			Lon:           b.Lon,
			LastHeartbeat: b.LastHeartbeat.Format(time.RFC3339),
			IsOnline:      isOnline,
			CurrentRideID: b.CurrentRideID,
		})
	}

	return res, nil
}

func (s *Service) SendCommand(ctx context.Context, slug, bikeID, command string) error {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return errors.New("campus tenant not found")
	}

	bike, err := s.bikeRepo.GetByID(ctx, bikeID)
	if err != nil || bike == nil {
		return errors.New("bike not found")
	}

	if bike.TenantID == nil || *bike.TenantID != tenant.ID {
		return errors.New("bike does not belong to this campus")
	}

	if s.mqttPub != nil {
		topic := fmt.Sprintf("bidkan/bikes/%s/commands", bikeID)
		payload := map[string]interface{}{
			"command":   command,
			"bike_id":   bikeID,
			"timestamp": time.Now().Unix(),
		}
		_ = s.mqttPub.Publish(topic, payload)
	}

	return nil
}

func (s *Service) ForceEndRide(ctx context.Context, slug, bikeID string) error {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return errors.New("campus tenant not found")
	}

	bike, err := s.bikeRepo.GetByID(ctx, bikeID)
	if err != nil || bike == nil {
		return errors.New("bike not found")
	}
	if bike.TenantID == nil || *bike.TenantID != tenant.ID {
		return errors.New("bike does not belong to this campus")
	}

	if bike.CurrentRideID != nil && s.rideRepo != nil {
		if rideUUID, parseErr := uuid.Parse(*bike.CurrentRideID); parseErr == nil {
			_ = s.rideRepo.EndRide(ctx, rideUUID, bike.Lat, bike.Lon, 0.0, "0.00", "0.00", "0.00")
		}
	}

	_ = s.bikeRepo.UpdateRideState(ctx, bikeID, "available", nil)

	if s.mqttPub != nil {
		topic := fmt.Sprintf("bidkan/bikes/%s/commands", bikeID)
		_ = s.mqttPub.Publish(topic, map[string]interface{}{
			"command":   "lock",
			"bike_id":   bikeID,
			"force":     true,
			"timestamp": time.Now().Unix(),
		})
	}

	return nil
}
