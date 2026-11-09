package end

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	domainService "github.com/Bangnus/Bidkan-backend/internal/app/domain/service"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/mqtt"
	"github.com/Bangnus/Bidkan-backend/internal/app/infrastructure/sqlc"
	"github.com/Bangnus/Bidkan-backend/pkg/utils/geo"
	"github.com/google/uuid"
)

type Request struct {
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	CouponCode string    `json:"coupon_code"`
}

type Service interface {
	End(ctx context.Context, req Request) (*entity.Ride, error)
}

type service struct {
	rideRepo     repository.RideRepository
	userRepo     repository.UserRepository
	bikeRepo     repository.BikeRepository
	zoneRepo     repository.ZoneRepository
	tenantRepo   repository.TenantRepository
	configRepo   repository.ConfigRepository
	spendingRepo repository.UserMonthlySpendingRepository
	rankCache    repository.UserRankCacheRepository
	configCache  repository.ConfigCacheRepository
	couponRepo   repository.CouponRepository
	mqttPub      mqtt.Publisher
	notiPub      domainService.NotificationProvider
	db           *sql.DB
}

func NewService(
	rideRepo repository.RideRepository,
	userRepo repository.UserRepository,
	bikeRepo repository.BikeRepository,
	zoneRepo repository.ZoneRepository,
	tenantRepo repository.TenantRepository,
	configRepo repository.ConfigRepository,
	spendingRepo repository.UserMonthlySpendingRepository,
	rankCache repository.UserRankCacheRepository,
	configCache repository.ConfigCacheRepository,
	couponRepo repository.CouponRepository,
	mqttPub mqtt.Publisher,
	notiPub domainService.NotificationProvider,
	db *sql.DB,
) Service {
	return &service{
		rideRepo:     rideRepo,
		userRepo:     userRepo,
		bikeRepo:     bikeRepo,
		zoneRepo:     zoneRepo,
		tenantRepo:   tenantRepo,
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
	now := time.Now()
	ride, err := s.rideRepo.GetActiveRide(ctx, req.UserID)
	if err != nil {
		return nil, errors.New("no active ride found for this user")
	}

	bike, err := s.bikeRepo.GetByID(ctx, ride.BikeID)
	if err != nil {
		return nil, errors.New("failed to retrieve bike information")
	}

	distance := geo.CalculateDistance(
		geo.Point{Lat: ride.StartLat, Lon: ride.StartLon},
		geo.Point{Lat: bike.Lat, Lon: bike.Lon},
	) / 1000.0

	if ride.Type == "service" {
		_ = s.rideRepo.EndRide(ctx, ride.ID, bike.Lat, bike.Lon, distance, "0.00", "0.00", "0.00")
		_ = s.bikeRepo.UpdateRideState(ctx, bike.BikeID, "available", nil)
		return s.rideRepo.GetByID(ctx, ride.ID)
	}

	tenant, err := s.tenantRepo.GetByID(ctx, ride.TenantID)
	if err != nil {
		tenant = &entity.Tenant{
			ID:                 ride.TenantID,
			UnlockFee:          parseOr(ride.UnlockFee, 5.0),
			PerMinuteRate:      parseOr(ride.RatePerMinute, 1.5),
			PlatformFeePercent: 18.0,
			ParkingPenaltyFee:  50.0,
		}
	}

	sinceYM, _ := strconv.Atoi(now.AddDate(0, -3, 0).Format("200601"))
	totalSpent, _ := s.spendingRepo.GetTotalInWindow(ctx, req.UserID, sinceYM)

	currentPoint := geo.Point{Lat: bike.Lat, Lon: bike.Lon}
	inParkingZone := s.checkCampusParkingZone(ctx, tenant.ID, currentPoint)

	var couponVal, couponMin float64
	var userCouponID *uuid.UUID
	queries := sqlc.New(s.db)
	if req.CouponCode != "" {
		if uc, err := queries.GetSpecificUserCouponByCode(ctx, sqlc.GetSpecificUserCouponByCodeParams{
			UserID: req.UserID,
			Code:   req.CouponCode,
		}); err == nil && (!uc.TenantID.Valid || uc.TenantID.UUID == tenant.ID) {
			couponVal, _ = strconv.ParseFloat(uc.Value, 64)
			couponMin, _ = strconv.ParseFloat(uc.MinAmount.String, 64)
			id := uc.UserCouponID
			userCouponID = &id
		}
	}

	seconds := now.Sub(ride.StartTime).Seconds()
	b := ComputeRideFare(tenant, seconds, totalSpent, inParkingZone, couponVal, couponMin)

	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	currentBalance, _ := strconv.ParseFloat(user.WalletBalance, 64)
	if err := s.userRepo.UpdateBalance(ctx, user.ID, fmt.Sprintf("%.2f", currentBalance-b.FinalFare)); err != nil {
		return nil, errors.New("failed to update user balance")
	}

	fareStr := fmt.Sprintf("%.2f", b.FinalFare)
	gpStr := fmt.Sprintf("%.2f", b.PlatformFee)
	netStr := fmt.Sprintf("%.2f", b.PartnerEarnings)

	if err := s.rideRepo.EndRide(ctx, ride.ID, bike.Lat, bike.Lon, distance, fareStr, gpStr, netStr); err != nil {
		return nil, errors.New("failed to end ride in database")
	}

	_ = s.bikeRepo.SaveLocation(ctx, *bike)
	_ = s.bikeRepo.UpdateRideState(ctx, bike.BikeID, "available", nil)

	if b.DiscountApplied > 0 && userCouponID != nil {
		_ = queries.MarkUserCouponAsUsed(ctx, *userCouponID)
	}

	s.notifyRideEnd(ctx, user, ride.ID, fareStr, distance, b.DiscountApplied)

	currentYM, _ := strconv.Atoi(now.Format("200601"))
	_ = s.spendingRepo.AddSpending(ctx, req.UserID, currentYM, b.FinalFare)
	_ = s.rankCache.DeleteRankCache(ctx, req.UserID.String())

	return s.rideRepo.GetByID(ctx, ride.ID)
}

func (s *service) checkCampusParkingZone(ctx context.Context, tenantID uuid.UUID, pt geo.Point) bool {
	zones, _ := s.zoneRepo.ListByTenantAndType(ctx, tenantID, "parking")
	legacyZones, _ := s.zoneRepo.ListByTenantAndType(ctx, tenantID, "P")
	all := append(zones, legacyZones...)
	if len(all) == 0 {
		// Wait until the campus operator creates parking points before penalizing riders.
		return true
	}
	return IsPointInAnyZone(pt, all)
}

func (s *service) notifyRideEnd(ctx context.Context, user *entity.User, rideID uuid.UUID, fareStr string, distance, discount float64) {
	if s.mqttPub != nil {
		s.mqttPub.Publish(fmt.Sprintf("bidkan/users/%s/notifications", user.ID), map[string]interface{}{
			"type":             "ride_end_summary",
			"ride_id":          rideID,
			"fare":             fareStr,
			"distance":         fmt.Sprintf("%.2f", distance),
			"discount_applied": discount,
			"time":             time.Now().Format(time.RFC3339),
		})
	}
	if s.notiPub != nil && user.FcmToken != "" {
		_ = s.notiPub.SendToToken(ctx, user.FcmToken, "สรุปการขี่จักรยาน", fmt.Sprintf("คุณใช้บริการเสร็จสิ้น ค่าบริการ %s บาท", fareStr), map[string]string{
			"type": "ride_end_summary",
		})
	}
}
