package tracking

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"context"
	"fmt"
)

type Service interface {
	ProcessTracking(ctx context.Context, data entity.BikeData) error
}

type service struct {
	bikeRepo  repository.BikeRepository
	bikeCache repository.BikeCacheRepository
}

func NewService(bikeRepo repository.BikeRepository, bikeCache repository.BikeCacheRepository) Service {
	return &service{
		bikeRepo:  bikeRepo,
		bikeCache: bikeCache,
	}
}

func (s *service) ProcessTracking(ctx context.Context, data entity.BikeData) error {
	// 1. อัปเดตสถานะล่าสุดในตาราง bikes (Current State)
	err := s.bikeRepo.UpdateStatus(ctx, data)
	if err != nil {
		fmt.Printf("❌ Failed to update bike status: %v\n", err)
		// ไม่ return error เพื่อให้ยังสามารถเซฟประวัติพิกัดต่อไปได้
	}

	// 2. บันทึกลงตาราง bike_locations (History)
	err = s.bikeRepo.SaveLocation(ctx, data)
	if err != nil {
		fmt.Printf("❌ Failed to save bike location history: %v\n", err)
		return err
	}

	// 3. อัปเดตลง Redis (Caching) เพื่อให้ดึงข้อมูลล่าสุดได้เร็ว
	err = s.bikeCache.SetLatestLocation(ctx, data)
	if err != nil {
		fmt.Printf("❌ Failed to cache bike location: %v\n", err)
	}

	fmt.Printf("📍 [อัปเดตพิกัด] รถ: %s | พิกัด: %f, %f | แบต: %d%% (DB & Redis)\n", 
		data.BikeID, data.Lat, data.Lon, data.Battery)
	
	return nil
}
