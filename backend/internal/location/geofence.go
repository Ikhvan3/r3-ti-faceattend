package location

import "math"

const earthRadiusMeters = 6371008.8

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

func ValidLatitude(value float64) bool {
	return finite(value) && value >= -90 && value <= 90
}

func ValidLongitude(value float64) bool {
	return finite(value) && value >= -180 && value <= 180
}

func ValidAccuracyMeters(value float64) bool {
	return finite(value) && value >= 0
}

func DistanceMeters(a Coordinate, b Coordinate) (float64, error) {
	if !ValidLatitude(a.Latitude) || !ValidLongitude(a.Longitude) || !ValidLatitude(b.Latitude) || !ValidLongitude(b.Longitude) {
		return 0, ErrInvalidInput
	}

	lat1 := degreesToRadians(a.Latitude)
	lat2 := degreesToRadians(b.Latitude)
	dLat := degreesToRadians(b.Latitude - a.Latitude)
	dLon := degreesToRadians(b.Longitude - a.Longitude)

	sinLat := math.Sin(dLat / 2)
	sinLon := math.Sin(dLon / 2)
	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLon*sinLon
	distance := 2 * earthRadiusMeters * math.Asin(math.Min(1, math.Sqrt(h)))
	if !finite(distance) {
		return 0, ErrInvalidInput
	}
	return distance, nil
}

func WithinRadius(center Coordinate, point Coordinate, radiusMeters int) (bool, float64, error) {
	if radiusMeters < 0 {
		return false, 0, ErrInvalidInput
	}
	distance, err := DistanceMeters(center, point)
	if err != nil {
		return false, 0, err
	}
	return distance <= float64(radiusMeters)+0.001, distance, nil
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
