package geo

import (
	"math"
	"strconv"
	"strings"
)

func ParsePLZ(input string) (string, error) {
	input = strings.TrimSpace(input)

	// Try coordinates
	if lat, lon, ok := parseCoordinates(input); ok {
		loc, err := FindNearest(lat, lon)
		if err != nil {
			return "", err
		}
		return padPLZ(loc.PLZ), nil
	}

	// Try numeric PLZ
	if isNumeric(input) {
		return padPLZ(input), nil
	}

	// Try name search
	loc, err := SearchLocation(input)
	if err != nil {
		return "", err
	}
	return padPLZ(loc.PLZ), nil
}

func padPLZ(plz string) string {
	for len(plz) < 6 {
		plz += "0"
	}
	return plz[:6]
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseCoordinates(s string) (lat, lon float64, ok bool) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	var err error
	lat, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, false
	}
	lon, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, false
	}
	return lat, lon, true
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
