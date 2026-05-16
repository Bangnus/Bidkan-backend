package entity

type SystemSummary struct {
	TotalUsers     int64  `json:"total_users"`
	TotalBikes     int64  `json:"total_bikes"`
	AvailableBikes int64  `json:"available_bikes"`
	ActiveRides    int64  `json:"active_rides"`
	TotalRevenue   string `json:"total_revenue"`
}

type BikeUsageStat struct {
	BikeID          string  `json:"bike_id"`
	RideCount       int64   `json:"ride_count"`
	TotalDistanceKM float64 `json:"total_distance_km"`
	TotalRevenue    string  `json:"total_revenue"`
}

type DailyRevenue struct {
	Date         string `json:"date"`
	RideCount    int64  `json:"ride_count"`
	DailyRevenue string `json:"daily_revenue"`
}

type FullReport struct {
	Summary      SystemSummary   `json:"summary"`
	BikeStats    []BikeUsageStat `json:"bike_stats"`
	DailyRevenue []DailyRevenue  `json:"daily_revenue"`
}
