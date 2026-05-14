package entity

type BikeData struct {
	BikeID  string  `json:"bike_id"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Battery int     `json:"battery"`
}
