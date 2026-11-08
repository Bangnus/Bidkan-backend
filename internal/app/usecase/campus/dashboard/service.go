package dashboard

import (
	"context"
	"errors"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	domainRepo "github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
)

type DashboardData struct {
	Tenant             *entity.Tenant `json:"tenant"`
	TotalBikes         int            `json:"total_bikes"`
	AvailableBikes     int            `json:"available_bikes"`
	InUseBikes         int            `json:"in_use_bikes"`
	LowBatteryBikes    int            `json:"low_battery_bikes"`
	TotalTripsToday    int            `json:"total_trips_today"`
	GrossVolumeToday   float64        `json:"gross_volume_today"`
	NetEarningsToday   float64        `json:"net_earnings_today"`
}

type Service struct {
	repo domainRepo.TenantRepository
}

func NewService(repo domainRepo.TenantRepository) *Service {
	return &Service{repo: repo}
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
	available := 0
	inUse := 0
	lowBattery := 0

	for _, b := range bikes {
		if b.Battery < 20 {
			lowBattery++
		}
		if b.Status == "available" {
			available++
		} else if b.Status == "in_use" {
			inUse++
		}
	}

	// Calculate simulated revenue metrics for the day
	tripsToday := total * 5
	grossVolume := float64(tripsToday) * 25.0
	partnerShare := (100.0 - tenant.PlatformFeePercent) / 100.0
	netEarnings := grossVolume * partnerShare

	return &DashboardData{
		Tenant:           tenant,
		TotalBikes:       total,
		AvailableBikes:   available,
		InUseBikes:       inUse,
		LowBatteryBikes:  lowBattery,
		TotalTripsToday:  tripsToday,
		GrossVolumeToday: grossVolume,
		NetEarningsToday: netEarnings,
	}, nil
}

