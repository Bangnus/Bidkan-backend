package end

import (
	"testing"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/pkg/utils/geo"
)

func TestComputeRideFareAndGPAttribution(t *testing.T) {
	kuTenant := &entity.Tenant{
		UnlockFee:          5.0,
		PerMinuteRate:      1.5,
		PlatformFeePercent: 18.0,
		ParkingPenaltyFee:  50.0,
	}

	// 10 minutes parked inside P zone, Bronze rank -> 5 + 15 = 20 THB, 18% GP = 3.60, partner = 16.40
	res := ComputeRideFare(kuTenant, 600, 0, true, 0, 0)
	if res.FinalFare != 20.00 || res.PlatformFee != 3.60 || res.PartnerEarnings != 16.40 {
		t.Fatalf("unexpected normal fare split: %+v", res)
	}

	// 10 minutes parked OUTSIDE P zone -> +50 penalty = 70 THB, 18% GP = 12.60, partner = 57.40
	out := ComputeRideFare(kuTenant, 600, 0, false, 0, 0)
	if out.PenaltyFee != 50.0 || out.FinalFare != 70.00 || out.PlatformFee != 12.60 || out.PartnerEarnings != 57.40 {
		t.Fatalf("unexpected outside-parking split: %+v", out)
	}

	// Gold rank (> 2500 spending -> -1 THB/min so effective 0.5/min) + Enterprise Tiered GP (10%)
	cmuEnterprise := &entity.Tenant{
		UnlockFee:          5.0,
		PerMinuteRate:      2.0,
		PlatformFeePercent: 10.0,
		ParkingPenaltyFee:  30.0,
	}
	gold := ComputeRideFare(cmuEnterprise, 600, 3000, true, 5.0, 0)
	// unlock 5 + 10*(1.0) = 15 - 5 coupon = 10 THB, 10% GP = 1.00, partner = 9.00
	if gold.EffectiveRate != 1.0 || gold.FinalFare != 10.0 || gold.PlatformFee != 1.0 || gold.PartnerEarnings != 9.0 {
		t.Fatalf("unexpected gold enterprise fare split: %+v", gold)
	}
}

func TestParseSingleAndArrayZoneBoundary(t *testing.T) {
	single := `{"lat":14.0234,"lon":99.9782}`
	pts := ParseZoneBoundary(single)
	if len(pts) != 1 || pts[0].Lat != 14.0234 {
		t.Fatalf("failed parsing single object boundary: %+v", pts)
	}
	z := []entity.Zone{{Boundary: single, Radius: 30}}
	if !IsPointInAnyZone(geo.Point{Lat: 14.02345, Lon: 99.97825}, z) {
		t.Fatalf("expected point inside single-object circle zone")
	}
}
