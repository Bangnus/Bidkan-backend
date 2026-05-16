package end

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/Bangnus/Bidkan-backend/pkg/utils/geo"
	"github.com/google/uuid"
)

type Request struct {
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	CouponCode string    `json:"coupon_code"` // เพิ่มฟิลด์เลือกใช้คูปอง (Optional)
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
	spendingRepo repository.UserMonthlySpendingRepository
	rankCache    repository.UserRankCacheRepository
	configCache  repository.ConfigCacheRepository
	couponRepo   repository.CouponRepository
	mqttPub      mqtt.Publisher
	notiPub      service.NotificationProvider
	db           *sql.DB // เพิ่ม DB เพื่อรัน Transaction ตอนจบงาน
}

func NewService(
	rideRepo repository.RideRepository,
	userRepo repository.UserRepository,
	bikeRepo repository.BikeRepository,
	zoneRepo repository.ZoneRepository,
	configRepo repository.ConfigRepository,
	spendingRepo repository.UserMonthlySpendingRepository,
	rankCache repository.UserRankCacheRepository,
	configCache repository.ConfigCacheRepository,
	couponRepo repository.CouponRepository,
	mqttPub mqtt.Publisher,
	notiPub service.NotificationProvider,
	db *sql.DB,
) Service {
	return &service{
		rideRepo:     rideRepo,
		userRepo:     userRepo,
		bikeRepo:     bikeRepo,
		zoneRepo:     zoneRepo,
		configRepo:   configRepo,
		spendingRepo: spendingRepo,
		rankCache:    rankCache,
		configCache:  configCache,
		couponRepo:   couponRepo,
		mqttPub:      mqttPub,
		notiPub:      notiPub,
		db:           db,
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

	// 2. คำนวณเวลาที่ใช้จริง
	now := time.Now()

	if ride.Type == "service" {
		// สำหรับงานบริการ/ซ่อมบำรุง ไม่คิดค่าใช้จ่าย และไม่เช็ค Geofencing
		distance := geo.CalculateDistance(geo.Point{Lat: ride.StartLat, Lon: ride.StartLon}, geo.Point{Lat: bike.Lat, Lon: bike.Lon}) / 1000.0
		_ = s.rideRepo.EndRide(ctx, ride.ID, bike.Lat, bike.Lon, distance, "0.00")

		bike.Status = "available" // หลังซ่อมเสร็จให้รถว่าง
		bike.CurrentRideID = nil
		_ = s.bikeRepo.UpdateStatus(ctx, *bike)

		return s.rideRepo.GetByID(ctx, ride.ID)
	}

	duration := now.Sub(ride.StartTime)
	seconds := duration.Seconds()

	// 3. คำนวณราคาตาม Rank (Real-time จากตารางสรุปยอด)
	// นับยอดสะสมย้อนหลัง 4 เดือน
	since := now.AddDate(0, -3, 0)
	sinceYM, _ := strconv.Atoi(since.Format("200601"))

	totalSpent, _ := s.spendingRepo.GetTotalInWindow(ctx, req.UserID, sinceYM)

	// กำหนดอัตราค่าบริการพื้นฐาน และส่วนลดตาม Rank
	baseRate := 2.0 // บาทต่อนาที
	discount := 0.0

	if totalSpent > 2500 {
		// Gold Rank
		discount = 1.0 // ลด 1 บาท
	} else if totalSpent > 800 {
		// Silver Rank
		discount = 0.5 // ลด 0.5 บาท
	}

	effectiveRate := baseRate - discount
	if effectiveRate < 0 {
		effectiveRate = 0
	}

	// คำนวณค่าเช่า (วินาทีละ effectiveRate/60 บาท)
	fareAmount := seconds * (effectiveRate / 60.0)

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

	// 5. คิดค่าปรับถ้าจอดนอกโซน P (ลองดึงจาก Cache ก่อน)
	if !inParkingZone {
		penaltyStr, err := s.configCache.GetConfigCache(ctx, "parking_penalty_fee")
		if err != nil {
			// ถ้าไม่มีใน Cache ให้ดึงจาก DB แล้วเก็บลง Cache (1ชั่วโมง)
			penaltyStr, err = s.configRepo.GetConfig(ctx, "parking_penalty_fee")
			if err == nil {
				_ = s.configCache.SetConfigCache(ctx, "parking_penalty_fee", penaltyStr, 1*time.Hour)
			}
		}

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

	// --- [NEW] เช็คคูปองส่วนลด (จะใช้ต่อเมื่อผู้ใช้ระบุโค้ดมาเท่านั้น) ---
	discountApplied := 0.0
	var userCouponID *uuid.UUID

	queries := sqlc.New(s.db)

	if req.CouponCode != "" {
		// ค้นหาคูปองเฉพาะเจาะจงที่ผู้ใช้ส่งมา (ต้องเป็นคูปองที่เก็บไว้แล้วและยังไม่ได้ใช้)
		uc, err := queries.GetSpecificUserCouponByCode(ctx, sqlc.GetSpecificUserCouponByCodeParams{
			UserID: req.UserID,
			Code:   req.CouponCode,
		})
		if err == nil {
			val, _ := strconv.ParseFloat(uc.Value, 64)
			min, _ := strconv.ParseFloat(uc.MinAmount.String, 64)

			if fareAmount >= min {
				discountApplied = val
				userCouponID = &uc.UserCouponID
				
				if discountApplied > fareAmount {
					discountApplied = fareAmount
				}
				fareAmount -= discountApplied
			}
		}
	}

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

	// 10. บันทึกแคชพิกัดล่าสุด (Optional)
	_ = s.bikeRepo.UpdateLocation(ctx, bike.ID, bike.Lat, bike.Lon)

	// --- [NEW] ส่งแจ้งเตือนจบงานผ่าน MQTT ---
	if s.mqttPub != nil {
		s.mqttPub.Publish(fmt.Sprintf("bidkan/users/%s/notifications", req.UserID), map[string]interface{}{
			"type":             "ride_end_summary",
			"ride_id":          ride.ID,
			"fare":             fareStr,
			"distance":         fmt.Sprintf("%.2f", distance),
			"discount_applied": discountApplied,
			"time":             time.Now().Format(time.RFC3339),
		})
	}

	// --- [NEW] ส่ง Push Notification ---
	if s.notiPub != nil {
		user, _ := s.userRepo.GetByID(ctx, req.UserID)
		if user != nil && user.FcmToken != "" {
			_ = s.notiPub.SendToToken(ctx, user.FcmToken, "สรุปการขี่จักรยาน", fmt.Sprintf("คุณใช้บริการเสร็จสิ้น ค่าบริการ %s บาท", fareStr), map[string]string{
				"type": "ride_end_summary",
			})
		}
	}

	// [NEW] มาร์คคูปองว่าใช้ไปแล้ว
	if userCouponID != nil {
		_ = queries.MarkUserCouponAsUsed(ctx, *userCouponID)
	}

	// 9. คืนสถานะรถจักรยาน
	bike.Status = "available"
	bike.CurrentRideID = nil
	_ = s.bikeRepo.UpdateStatus(ctx, *bike)

	// 10. อัปเดตตารางสรุปยอดรายเดือนเพื่อความรวดเร็วในการคำนวณครั้งถัดไป
	currentYM, _ := strconv.Atoi(now.Format("200601"))
	_ = s.spendingRepo.AddSpending(ctx, req.UserID, currentYM, fareAmount)

	// 11. ลบ Cache ของ Rank เพื่อให้การเปิดดูครั้งหน้าคำนวณใหม่ (เพราะยอดเงินเปลี่ยนแล้ว)
	_ = s.rankCache.DeleteRankCache(ctx, req.UserID.String())

	return s.rideRepo.GetByID(ctx, ride.ID)
}
