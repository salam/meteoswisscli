package geo

import (
	"fmt"
	"strconv"
	"strings"
)

func ParsePLZ(input string) (string, error) {
	input = strings.TrimSpace(input)

	if _, _, ok := parseCoordinates(input); ok {
		return "", fmt.Errorf("coordinate lookup not yet supported, use a PLZ code")
	}

	if isNumeric(input) {
		return padPLZ(input), nil
	}

	return "", fmt.Errorf("location not found. Try a PLZ code (e.g. 8001) or place name")
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
