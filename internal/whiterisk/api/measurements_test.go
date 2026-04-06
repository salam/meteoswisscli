package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIMISStations(t *testing.T) {
	stations := []IMISStation{
		{Code: "DAV2", Label: "Davos", Lon: 9.85, Lat: 46.81, Elevation: 2560, CountryCode: "CH", CantonCode: "GR", Type: "SNOW_FLAT"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(stations)
	}))
	defer server.Close()

	c := NewClientWithBase(server.URL, "en")
	result, err := c.GetIMISStations()
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(result) != 1 || result[0].Code != "DAV2" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestGetIMISMeasurements(t *testing.T) {
	measurements := []IMISMeasurement{
		{StationCode: "DAV2", MeasureDate: "2026-04-05T18:00:00Z", HS: floatPtr(142), TA30MinMean: floatPtr(-2.3)},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(measurements)
	}))
	defer server.Close()

	c := NewClientWithBase(server.URL, "en")
	result, err := c.GetIMISMeasurementsByStation("DAV2", 1)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(result) != 1 || result[0].StationCode != "DAV2" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func floatPtr(f float64) *float64 { return &f }
