package location

import (
	"errors"
	"math"
	"testing"
)

func TestGeofenceCalculator(t *testing.T) {
	center := Coordinate{Latitude: -6.2, Longitude: 106.816666}

	t.Run("titik sama mendekati nol", func(t *testing.T) {
		distance, err := DistanceMeters(center, center)
		if err != nil {
			t.Fatalf("DistanceMeters() error = %v", err)
		}
		if distance > 0.001 {
			t.Fatalf("distance = %f", distance)
		}
	})

	t.Run("di dalam radius", func(t *testing.T) {
		ok, _, err := WithinRadius(center, Coordinate{Latitude: -6.2001, Longitude: 106.816666}, 20)
		if err != nil || !ok {
			t.Fatalf("WithinRadius() ok=%v err=%v", ok, err)
		}
	})

	t.Run("sekitar batas radius", func(t *testing.T) {
		ok, distance, err := WithinRadius(center, Coordinate{Latitude: -6.2009, Longitude: 106.816666}, 101)
		if err != nil || !ok {
			t.Fatalf("WithinRadius() ok=%v distance=%f err=%v", ok, distance, err)
		}
	})

	t.Run("di luar radius", func(t *testing.T) {
		ok, _, err := WithinRadius(center, Coordinate{Latitude: -6.202, Longitude: 106.816666}, 100)
		if err != nil {
			t.Fatalf("WithinRadius() error = %v", err)
		}
		if ok {
			t.Fatalf("point should be outside radius")
		}
	})

	t.Run("invalid input ditolak", func(t *testing.T) {
		if ValidLatitude(91) || ValidLongitude(181) || ValidAccuracyMeters(math.NaN()) || ValidLatitude(math.Inf(1)) {
			t.Fatalf("invalid coordinate accepted")
		}
		_, err := DistanceMeters(Coordinate{Latitude: math.NaN(), Longitude: 1}, center)
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidInput)
		}
	})

	t.Run("simetris", func(t *testing.T) {
		a := Coordinate{Latitude: -6.2, Longitude: 106.8}
		b := Coordinate{Latitude: -6.21, Longitude: 106.82}
		ab, err := DistanceMeters(a, b)
		if err != nil {
			t.Fatalf("DistanceMeters(a,b) error = %v", err)
		}
		ba, err := DistanceMeters(b, a)
		if err != nil {
			t.Fatalf("DistanceMeters(b,a) error = %v", err)
		}
		if math.Abs(ab-ba) > 0.001 {
			t.Fatalf("distance not symmetric: %f vs %f", ab, ba)
		}
	})
}
