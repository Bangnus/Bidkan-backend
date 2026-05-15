package geo

import (
	"math"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// IsPointInPolygon uses the Ray-casting algorithm to determine if a point is inside a polygon
func IsPointInPolygon(point Point, polygon []Point) bool {
	inside := false
	j := len(polygon) - 1
	for i := 0; i < len(polygon); i++ {
		if (polygon[i].Lat < point.Lat && polygon[j].Lat >= point.Lat || polygon[j].Lat < point.Lat && polygon[i].Lat >= point.Lat) {
			if polygon[i].Lon+(point.Lat-polygon[i].Lat)/(polygon[j].Lat-polygon[i].Lat)*(polygon[j].Lon-polygon[i].Lon) < point.Lon {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

// CalculateDistance uses the Haversine formula to calculate the distance between two points in meters
func CalculateDistance(p1, p2 Point) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (p2.Lat - p1.Lat) * (math.Pi / 180.0)
	dLon := (p2.Lon - p1.Lon) * (math.Pi / 180.0)

	lat1 := p1.Lat * (math.Pi / 180.0)
	lat2 := p2.Lat * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// IsPointInCircle checks if a point is within a given radius (in meters) of a center point
func IsPointInCircle(point, center Point, radiusMeters float64) bool {
	distance := CalculateDistance(point, center)
	return distance <= radiusMeters
}
