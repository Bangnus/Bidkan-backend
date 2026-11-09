package end

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/pkg/utils/geo"
)

type FareBreakdown struct {
	BaseRatePerMin  float64
	EffectiveRate   float64
	UnlockFee       float64
	TimeFare        float64
	PenaltyFee      float64
	DiscountApplied float64
	FinalFare       float64
	PlatformFee     float64
	PartnerEarnings float64
}

func ParseZoneBoundary(boundary string) []geo.Point {
	var pts []geo.Point
	if err := json.Unmarshal([]byte(boundary), &pts); err == nil && len(pts) > 0 {
		return pts
	}
	var single geo.Point
	if err := json.Unmarshal([]byte(boundary), &single); err == nil && (single.Lat != 0 || single.Lon != 0) {
		return []geo.Point{single}
	}
	return nil
}

func IsPointInAnyZone(pt geo.Point, zones []entity.Zone) bool {
	for _, z := range zones {
		pts := ParseZoneBoundary(z.Boundary)
		if len(pts) == 0 {
			continue
		}
		if z.Radius > 0 {
			if geo.IsPointInCircle(pt, pts[0], z.Radius) {
				return true
			}
		} else if geo.IsPointInPolygon(pt, pts) {
			return true
		}
	}
	return false
}

func ComputeRideFare(
	tenant *entity.Tenant,
	seconds, totalSpentWindow float64,
	inParkingZone bool,
	couponValue, couponMinAmount float64,
) FareBreakdown {
	unlockFee := tenant.UnlockFee
	baseRate := tenant.PerMinuteRate
	if baseRate <= 0 {
		baseRate = 1.5
	}

	rankDiscount := 0.0
	if totalSpentWindow > 2500 {
		rankDiscount = 1.0
	} else if totalSpentWindow > 800 {
		rankDiscount = 0.5
	}

	effectiveRate := baseRate - rankDiscount
	if effectiveRate < 0 {
		effectiveRate = 0
	}

	timeFare := seconds * (effectiveRate / 60.0)
	penalty := 0.0
	if !inParkingZone && tenant.ParkingPenaltyFee > 0 {
		penalty = tenant.ParkingPenaltyFee
	}

	gross := math.Round((unlockFee+timeFare+penalty)*100) / 100
	discountApplied := 0.0
	if couponValue > 0 && gross >= couponMinAmount {
		discountApplied = couponValue
		if discountApplied > gross {
			discountApplied = gross
		}
		discountApplied = math.Round(discountApplied*100) / 100
	}

	finalFare := math.Round((gross-discountApplied)*100) / 100
	gpPercent := tenant.PlatformFeePercent
	if gpPercent < 0 {
		gpPercent = 0
	}
	platformFee := math.Round((finalFare*(gpPercent/100.0))*100) / 100
	partnerEarnings := math.Round((finalFare-platformFee)*100) / 100

	return FareBreakdown{
		BaseRatePerMin:  baseRate,
		EffectiveRate:   effectiveRate,
		UnlockFee:       unlockFee,
		TimeFare:        math.Round(timeFare*100) / 100,
		PenaltyFee:      penalty,
		DiscountApplied: discountApplied,
		FinalFare:       finalFare,
		PlatformFee:     platformFee,
		PartnerEarnings: partnerEarnings,
	}
}

func parseOr(raw string, fallback float64) float64 {
	if v, err := strconv.ParseFloat(raw, 64); err == nil && v >= 0 {
		return v
	}
	return fallback
}
