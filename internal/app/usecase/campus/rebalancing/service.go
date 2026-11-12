package rebalancing

import (
	"context"
	"encoding/json"
	"errors"
	"math"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type Service struct {
	tenantRepo domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
	rebalRepo  domainRepo.RebalancingRepository
}

func NewService(
	tenantRepo domainRepo.TenantRepository,
	campusRepo domainRepo.CampusOpsRepository,
	rebalRepo domainRepo.RebalancingRepository,
) *Service {
	return &Service{
		tenantRepo: tenantRepo,
		campusRepo: campusRepo,
		rebalRepo:  rebalRepo,
	}
}

func parseCenter(boundary string) (float64, float64) {
	type coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
	var arr []coord
	if err := json.Unmarshal([]byte(boundary), &arr); err == nil && len(arr) > 0 {
		return arr[0].Lat, arr[0].Lon
	}
	var c coord
	if err := json.Unmarshal([]byte(boundary), &c); err == nil {
		return c.Lat, c.Lon
	}
	return 0, 0
}

func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (s *Service) GetRebalanceOverview(ctx context.Context, slug string) (*entity.RebalanceOverviewResponse, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	zones, err := s.campusRepo.ListZonesByTenant(ctx, tenant.ID)
	if err != nil {
		return nil, err
	}

	bikes, err := s.tenantRepo.ListBikesByTenant(ctx, tenant.ID)
	if err != nil {
		return nil, err
	}

	res := &entity.RebalanceOverviewResponse{
		Zones:         make([]*entity.ZoneBalanceStatus, 0, len(zones)),
		DispatchTasks: make([]*entity.RebalanceTaskAdvice, 0),
	}

	overflowList := make([]*entity.ZoneBalanceStatus, 0)
	deficitList := make([]*entity.ZoneBalanceStatus, 0)

	for _, z := range zones {
		lat, lon := parseCenter(z.Boundary)
		if lat == 0 && lon == 0 {
			continue
		}

		count := 0
		for _, b := range bikes {
			if b.Lat != 0 && b.Lon != 0 {
				if distanceMeters(lat, lon, b.Lat, b.Lon) <= z.Radius {
					count++
				}
			}
		}

		target := 8
		if z.Type == "battery_hub" {
			target = 15
		}

		status := "balanced"
		delta := count - target
		if delta >= 3 {
			status = "overflow"
			res.OverflowZonesCount++
		} else if delta <= -3 {
			status = "deficit"
			res.DeficitZonesCount++
		}

		zb := &entity.ZoneBalanceStatus{
			ZoneID:       z.ID,
			ZoneName:     z.Name,
			ZoneType:     z.Type,
			Lat:          lat,
			Lon:          lon,
			Radius:       z.Radius,
			CurrentBikes: count,
			TargetBikes:  target,
			Status:       status,
			Delta:        delta,
		}
		res.Zones = append(res.Zones, zb)
		res.TotalBikesInZones += count

		if status == "overflow" {
			overflowList = append(overflowList, zb)
		} else if status == "deficit" {
			deficitList = append(deficitList, zb)
		}
	}

	// Generate smart dispatch advice matching overflow to deficit
	for _, over := range overflowList {
		for _, def := range deficitList {
			suggested := int(math.Min(float64(over.Delta), float64(-def.Delta)))
			if suggested > 0 {
				priority := "medium"
				if suggested >= 5 {
					priority = "high"
				}
				res.DispatchTasks = append(res.DispatchTasks, &entity.RebalanceTaskAdvice{
					FromZoneName:   over.ZoneName,
					ToZoneName:     def.ZoneName,
					SuggestedBikes: suggested,
					Priority:       priority,
				})
				over.Delta -= suggested
				def.Delta += suggested
			}
		}
	}

	heatmap, _ := s.rebalRepo.GetHeatmapPoints(ctx, tenant.ID, 400)
	res.HeatmapPoints = heatmap

	return res, nil
}
