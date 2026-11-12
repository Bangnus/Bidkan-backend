package entity

import "github.com/google/uuid"

type ZoneBalanceStatus struct {
	ZoneID       uuid.UUID `json:"zone_id"`
	ZoneName     string    `json:"zone_name"`
	ZoneType     string    `json:"zone_type"`
	Lat          float64   `json:"lat"`
	Lon          float64   `json:"lon"`
	Radius       float64   `json:"radius"`
	CurrentBikes int       `json:"current_bikes"`
	TargetBikes  int       `json:"target_bikes"`
	Status       string    `json:"status"` // overflow, deficit, balanced
	Delta        int       `json:"delta"`  // CurrentBikes - TargetBikes
}

type RebalanceTaskAdvice struct {
	FromZoneName   string `json:"from_zone_name"`
	ToZoneName     string `json:"to_zone_name"`
	SuggestedBikes int    `json:"suggested_bikes"`
	Priority       string `json:"priority"` // high, medium, low
}

type HeatmapPoint struct {
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Intensity float64 `json:"intensity"`
	Type      string  `json:"type"` // pickup, dropoff
}

type RebalanceOverviewResponse struct {
	TotalBikesInZones  int                    `json:"total_bikes_in_zones"`
	OverflowZonesCount int                    `json:"overflow_zones_count"`
	DeficitZonesCount  int                    `json:"deficit_zones_count"`
	Zones              []*ZoneBalanceStatus   `json:"zones"`
	DispatchTasks      []*RebalanceTaskAdvice `json:"dispatch_tasks"`
	HeatmapPoints      []*HeatmapPoint        `json:"heatmap_points"`
}

