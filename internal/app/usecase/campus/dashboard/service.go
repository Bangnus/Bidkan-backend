package dashboard

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type DashboardData struct {
	Tenant           *entity.Tenant    `json:"tenant"`
	GPTier           entity.GPTierInfo `json:"gp_tier"`
	TotalBikes       int               `json:"total_bikes"`
	AvailableBikes   int               `json:"available_bikes"`
	InUseBikes       int               `json:"in_use_bikes"`
	LowBatteryBikes  int               `json:"low_battery_bikes"`
	TotalTripsToday  int               `json:"total_trips_today"`
	GrossVolumeToday float64           `json:"gross_volume_today"`
	NetEarningsToday float64           `json:"net_earnings_today"`
}

type Service struct {
	repo       domainRepo.TenantRepository
	campusRepo domainRepo.CampusOpsRepository
}

func NewService(repo domainRepo.TenantRepository, campusRepo domainRepo.CampusOpsRepository) *Service {
	return &Service{repo: repo, campusRepo: campusRepo}
}

func (s *Service) Execute(ctx context.Context, slug string) (*DashboardData, error) {
	tenant, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("campus tenant not found")
	}

	bikes, err := s.repo.ListBikesByTenant(ctx, tenant.ID)
	if err != nil {
		bikes = []*entity.BikeData{}
	}

	total := len(bikes)
	available, inUse, lowBattery := 0, 0, 0
	for _, b := range bikes {
		if b.Battery < 20 {
			lowBattery++
		}
		switch b.Status {
		case "available":
			available++
		case "in_use":
			inUse++
		}
	}

	todayStats := &entity.RideStatsSummary{}
	if s.campusRepo != nil {
		if st, err := s.campusRepo.GetTodayStats(ctx, tenant.ID); err == nil && st != nil {
			todayStats = st
		}
	}

	fleetCount := total
	if cnt, err := s.repo.CountFleetForTenant(ctx, tenant.ID); err == nil && cnt > 0 {
		fleetCount = cnt
	}

	return &DashboardData{
		Tenant:           tenant,
		GPTier:           entity.CalculateTieredGP(fleetCount),
		TotalBikes:       total,
		AvailableBikes:   available,
		InUseBikes:       inUse,
		LowBatteryBikes:  lowBattery,
		TotalTripsToday:  todayStats.TotalRides,
		GrossVolumeToday: todayStats.GrossAmount,
		NetEarningsToday: todayStats.NetAmount,
	}, nil
}
