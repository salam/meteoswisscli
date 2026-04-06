package geo

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed data/meteoswiss_stations.json
var stationsData []byte

type Station struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Elevation int     `json:"elevation"`
	Canton    string  `json:"canton"`
}

type StationWithDist struct {
	Station  Station
	Distance float64 // km
}

var meteoStations []Station

func loadStations() error {
	if meteoStations != nil {
		return nil
	}
	return json.Unmarshal(stationsData, &meteoStations)
}

// FindNearestStations returns stations sorted by distance from the given coordinates.
func FindNearestStations(lat, lon float64, limit int) ([]StationWithDist, error) {
	if err := loadStations(); err != nil {
		return nil, fmt.Errorf("load stations: %w", err)
	}

	var result []StationWithDist
	for _, s := range meteoStations {
		d := haversineDistance(lat, lon, s.Lat, s.Lon)
		result = append(result, StationWithDist{Station: s, Distance: d})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Distance < result[j].Distance })

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// ResolveStation takes a station code, place name, or lat,lon and returns
// the nearest station(s). If input is a station code, returns that station.
func ResolveStation(input string, limit int) ([]StationWithDist, error) {
	if err := loadStations(); err != nil {
		return nil, fmt.Errorf("load stations: %w", err)
	}

	input = strings.TrimSpace(input)

	// Try exact station code match
	upper := strings.ToUpper(input)
	for _, s := range meteoStations {
		if strings.ToUpper(s.Code) == upper {
			return []StationWithDist{{Station: s, Distance: 0}}, nil
		}
	}

	// Try coordinates
	if lat, lon, ok := parseCoordinates(input); ok {
		return FindNearestStations(lat, lon, limit)
	}

	// Try place name → resolve to coords → find nearest stations
	loc, err := SearchLocation(input)
	if err != nil {
		return nil, err
	}
	return FindNearestStations(loc.Lat, loc.Lon, limit)
}
