package end

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/pkg/utils/geo"
	"github.com/google/uuid"
)

type Request struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

type Service interface {
	End(ctx context.Context, req Request) (*entity.Ride, error)
}

type service struct {
	rideRepo   repository.RideRepository
	userRepo   repository.UserRepository
	bikeRepo   repository.BikeRepository
	zoneRepo   repository.ZoneRepository
	configRepo repository.ConfigRepository
}

func NewService(rideRepo repository.RideRepository, userRepo repository.UserRepository, bikeRepo repository.BikeRepository, zoneRepo repository.ZoneRepository, configRepo repository.ConfigRepository) Service {
	return &service{
		rideRepo:   rideRepo,
		userRepo:   userRepo,
		bikeRepo:   bikeRepo,
		zoneRepo:   zoneRepo,
		configRepo: configRepo,
	}
}

func (s *service) End(ctx context.Context, req Request) (*entity.Ride, error) {
	// 1. หาการเช่าที่กำลังดำเนินอยู่ (Active Ride)
	ride, err := s.rideRepo.GetActiveRide(ctx, req.UserID)
	if err != nil {
		return nil, errors.New("no active ride found for this user")
	}

	// ดึงข้อมูลรถจักรยานเพื่อใช้พิกัดล่าสุดที่อัปเดตจาก MQTT
	bike, err := s.bikeRepo.GetByID(ctx, ride.BikeID)
	if err != nil {
		return nil, errors.New("failed to retrieve bike information")
	}

	// 2. คำนวณเวลาที่ใช้จริงเป็น "วินาที" เพื่อความแม่นยำระดับสตางค์
	now := time.Now()
	duration := now.Sub(ride.StartTime)
	seconds := duration.Seconds()

	// 3. คำนวณค่าเช่าปกติ (นาทีละ 2 บาท => วินาทีละ 2/60 บาท)
	fareAmount := seconds * (2.0 / 60.0)

	// 4. ตรวจสอบว่าจอดในโซน P หรือไม่ (Geofencing) อ้างอิงจากพิกัดรถจักรยาน
	inParkingZone := false
	parkingZones, _ := s.zoneRepo.ListByType(ctx, "P")

	currentPoint := geo.Point{Lat: bike.Lat, Lon: bike.Lon}

	for _, zone := range parkingZones {
		var boundary []geo.Point
		if err := json.Unmarshal([]byte(zone.Boundary), &boundary); err == nil && len(boundary) > 0 {
			if zone.Radius > 0 {
				// เช็คแบบวงกลม (ใช้พิกัดแรกใน boundary เป็นจุดศูนย์กลาง)
				if geo.IsPointInCircle(currentPoint, boundary[0], zone.Radius) {
					inParkingZone = true
					break
				}
			} else {
				// เช็คแบบรูปหลายเหลี่ยม (Polygon)
				if geo.IsPointInPolygon(currentPoint, boundary) {
					inParkingZone = true
					break
				}
			}
		}
	}

	// 5. คิดค่าปรับถ้าจอดนอกโซน P
	if !inParkingZone {
		penaltyStr, err := s.configRepo.GetConfig(ctx, "parking_penalty_fee")
		penaltyFee := 50.0 // ค่าเริ่มต้นถ้าไม่พบการตั้งค่า
		if err == nil {
			if p, err := strconv.ParseFloat(penaltyStr, 64); err == nil {
				penaltyFee = p
			}
		}
		fareAmount += penaltyFee // บวกค่าปรับเข้าไป
	}

	// ปัดเศษให้เหลือ 2 ตำแหน่งสำหรับเก็บลง Database (เช่น 4.33 บาท)
	fareAmount = math.Round(fareAmount*100) / 100
	fareStr := fmt.Sprintf("%.2f", fareAmount)

	// 6. หักเงินใน Wallet
	user, _ := s.userRepo.GetByID(ctx, req.UserID)
	currentBalance, _ := strconv.ParseFloat(user.WalletBalance, 64)

	newBalance := currentBalance - fareAmount
	newBalanceStr := fmt.Sprintf("%.2f", newBalance)

	// อัปเดตยอดเงินในฐานข้อมูล
	err = s.userRepo.UpdateBalance(ctx, user.ID, newBalanceStr)
	if err != nil {
		return nil, errors.New("failed to update user balance")
	}

	// 7. อัปเดตระยะทางเบื้องต้น
	distance := geo.CalculateDistance(geo.Point{Lat: ride.StartLat, Lon: ride.StartLon}, currentPoint) / 1000.0 // Convert to km

	// 8. อัปเดตสถานะการเช่าเป็น completed โดยใช้พิกัดจากจักรยาน
	err = s.rideRepo.EndRide(ctx, ride.ID, bike.Lat, bike.Lon, distance, fareStr)
	if err != nil {
		return nil, errors.New("failed to end ride in database")
	}

	// 9. คืนสถานะรถจักรยาน
	bike.Status = "available"
	bike.CurrentRideID = nil
	_ = s.bikeRepo.UpdateStatus(ctx, *bike)

	return s.rideRepo.GetByID(ctx, ride.ID)
}
