package entity

type GPTierInfo struct {
	FleetCount       int     `json:"fleet_count"`
	CurrentGPPercent float64 `json:"current_gp_percent"`
	TierName         string  `json:"tier_name"`
	NextTierTarget   int     `json:"next_tier_target"`
	NextTierGP       float64 `json:"next_tier_gp"`
	BikesToNextTier  int     `json:"bikes_to_next_tier"`
}

// CalculateTieredGP computes BidKan GP fee percentage by active/owned fleet volume:
//   - 0-30 bikes   -> 18% (Starter Tier)
//   - 31-100 bikes -> 15% (Growth Tier)
//   - 101+ bikes   -> 10% (Enterprise Tier)
func CalculateTieredGP(fleetCount int) GPTierInfo {
	if fleetCount > 100 {
		return GPTierInfo{
			FleetCount:       fleetCount,
			CurrentGPPercent: 10.0,
			TierName:         "Enterprise (101+ คัน)",
			NextTierTarget:   0,
			NextTierGP:       10.0,
			BikesToNextTier:  0,
		}
	}
	if fleetCount >= 31 {
		return GPTierInfo{
			FleetCount:       fleetCount,
			CurrentGPPercent: 15.0,
			TierName:         "Growth (31-100 คัน)",
			NextTierTarget:   101,
			NextTierGP:       10.0,
			BikesToNextTier:  101 - fleetCount,
		}
	}
	needed := 31 - fleetCount
	if needed < 0 {
		needed = 31
	}
	return GPTierInfo{
		FleetCount:       fleetCount,
		CurrentGPPercent: 18.0,
		TierName:         "Starter (1-30 คัน)",
		NextTierTarget:   31,
		NextTierGP:       15.0,
		BikesToNextTier:  needed,
	}
}
