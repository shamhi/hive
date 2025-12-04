package utils

import (
	"hive/services/dispatch/internal/domain/shared"
	"math"
)

func Dist(p1, p2 *shared.Location) float64 {
	const R = 6371e3

	lat1Rad := p1.Lat * math.Pi / 180
	lat2Rad := p2.Lat * math.Pi / 180
	deltaLatRad := (p2.Lat - p1.Lat) * math.Pi / 180
	deltaLonRad := (p2.Lon - p1.Lon) * math.Pi / 180

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
