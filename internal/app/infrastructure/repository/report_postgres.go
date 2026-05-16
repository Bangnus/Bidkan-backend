package repository

import (
	"context"
	"database/sql"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
)

type reportPostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewReportPostgresRepository(db *sql.DB) repository.ReportRepository {
	return &reportPostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *reportPostgresRepository) GetSummary(ctx context.Context) (*entity.FullReport, error) {
	// 1. Get General Summary
	summaryRes, err := r.queries.GetSystemSummary(ctx)
	if err != nil {
		return nil, err
	}

	summary := entity.SystemSummary{
		TotalUsers:     summaryRes.TotalUsers,
		TotalBikes:     summaryRes.TotalBikes,
		AvailableBikes: summaryRes.AvailableBikes,
		ActiveRides:    summaryRes.ActiveRides,
		TotalRevenue:   summaryRes.TotalRevenue,
	}

	// 2. Get Bike Usage Stats
	bikeRes, err := r.queries.GetBikeUsageStats(ctx)
	if err != nil {
		return nil, err
	}

	bikeStats := make([]entity.BikeUsageStat, 0)
	for _, b := range bikeRes {
		bikeStats = append(bikeStats, entity.BikeUsageStat{
			BikeID:          b.BikeID,
			RideCount:       b.RideCount,
			TotalDistanceKM: b.TotalDistanceKm,
			TotalRevenue:    b.TotalRevenue,
		})
	}

	// 3. Get Daily Revenue
	dailyRes, err := r.queries.GetDailyRevenue(ctx)
	if err != nil {
		return nil, err
	}

	dailyRevenue := make([]entity.DailyRevenue, 0)
	for _, d := range dailyRes {
		// d.Date เป็น time.Time จาก sqlc อยู่แล้ว
		dateStr := d.Date.Format("2006-01-02")

		dailyRevenue = append(dailyRevenue, entity.DailyRevenue{
			Date:         dateStr,
			RideCount:    d.RideCount,
			DailyRevenue: d.DailyRevenue,
		})
	}

	return &entity.FullReport{
		Summary:      summary,
		BikeStats:    bikeStats,
		DailyRevenue: dailyRevenue,
	}, nil
}
