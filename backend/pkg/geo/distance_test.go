package geo

import "testing"

func TestHaversineMetersZeroDistance(t *testing.T) {
	d := HaversineMeters(10.762622, 106.660172, 10.762622, 106.660172)
	if d > 1e-6 {
		t.Errorf("expected ~0 distance, got %f", d)
	}
}

func TestHaversineMetersOneDegreeLatitude(t *testing.T) {
	// One degree of latitude is ~111.19km everywhere on the WGS84 ellipsoid.
	d := HaversineMeters(0, 0, 1, 0)
	if d < 110900 || d > 111400 {
		t.Errorf("expected ~111195m, got %f", d)
	}
}

func TestHaversineMetersMatchesRestrictedZoneRadius(t *testing.T) {
	// Point known to sit just inside the 800m Tan Son Nhat NFZ radius used
	// by the suggestion engine (proto/suggestion.proto + main.go RTH check).
	nfzLat, nfzLng := 10.7725, 106.69
	d := HaversineMeters(nfzLat, nfzLng, 10.7725, 106.6965)
	if d >= 800.0 {
		t.Errorf("expected point inside 800m NFZ radius, got %f meters", d)
	}
}
