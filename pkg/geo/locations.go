package geo

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed data/plz_locations.json
var locationsData []byte

type Location struct {
	PLZ      string  `json:"plz"`
	Name     string  `json:"name"`
	Gemeinde string  `json:"gemeinde"`
	Kanton   string  `json:"kanton"`
	Lon      float64 `json:"lon"`
	Lat      float64 `json:"lat"`
}

var locations []Location

func loadLocations() error {
	if locations != nil {
		return nil
	}
	return json.Unmarshal(locationsData, &locations)
}

func SearchLocation(query string) (*Location, error) {
	if err := loadLocations(); err != nil {
		return nil, fmt.Errorf("load locations: %w", err)
	}

	query = strings.TrimSpace(query)
	lower := strings.ToLower(query)

	// Parse "Name KT" format like "Arosa GR"
	var kantonFilter string
	parts := strings.Fields(query)
	if len(parts) >= 2 {
		last := strings.ToUpper(parts[len(parts)-1])
		if len(last) == 2 {
			kantonFilter = last
			lower = strings.ToLower(strings.Join(parts[:len(parts)-1], " "))
		}
	}

	// Exact name match (Ortschaftsname) first
	for i := range locations {
		l := &locations[i]
		if strings.ToLower(l.Name) == lower && (kantonFilter == "" || l.Kanton == kantonFilter) {
			return l, nil
		}
	}

	// Exact gemeinde match
	for i := range locations {
		l := &locations[i]
		if strings.ToLower(l.Gemeinde) == lower && (kantonFilter == "" || l.Kanton == kantonFilter) {
			return l, nil
		}
	}

	// Prefix match on name
	for i := range locations {
		l := &locations[i]
		if strings.HasPrefix(strings.ToLower(l.Name), lower) && (kantonFilter == "" || l.Kanton == kantonFilter) {
			return l, nil
		}
	}

	// Prefix match on gemeinde
	for i := range locations {
		l := &locations[i]
		if strings.HasPrefix(strings.ToLower(l.Gemeinde), lower) && (kantonFilter == "" || l.Kanton == kantonFilter) {
			return l, nil
		}
	}

	// Contains match
	for i := range locations {
		l := &locations[i]
		nameMatch := strings.Contains(strings.ToLower(l.Name), lower) ||
			strings.Contains(strings.ToLower(l.Gemeinde), lower)
		if nameMatch && (kantonFilter == "" || l.Kanton == kantonFilter) {
			return l, nil
		}
	}

	return nil, fmt.Errorf("location not found. Try a PLZ code (e.g. 8001) or place name")
}

func FindNearest(lat, lon float64) (*Location, error) {
	if err := loadLocations(); err != nil {
		return nil, fmt.Errorf("load locations: %w", err)
	}

	var best *Location
	bestDist := 1e18
	for i := range locations {
		l := &locations[i]
		d := haversineDistance(lat, lon, l.Lat, l.Lon)
		if d < bestDist {
			bestDist = d
			best = l
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no locations found")
	}
	return best, nil
}
