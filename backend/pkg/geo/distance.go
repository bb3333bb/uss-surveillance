// Package geo provides shared WGS84 coordinate math used across the gateway.
package geo

import "math"

const earthRadiusMeters = 6371000.0

// HaversineMeters computes the great-circle distance in meters between two
// WGS84 lat/lng points.
func HaversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}
