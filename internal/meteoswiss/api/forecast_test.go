package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetForecast(t *testing.T) {
	response := PlzDetail{
		CurrentWeather: CurrentWeather{
			Time:        1775446800000,
			Icon:        1,
			Temperature: 16.8,
		},
		Forecast: []ForecastDay{
			{DayDate: "2026-04-05", IconDay: 1, TemperatureMax: 24, TemperatureMin: 10},
			{DayDate: "2026-04-06", IconDay: 2, TemperatureMax: 20, TemperatureMin: 8},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/plzDetail" {
			t.Errorf("path = %q, want /v1/plzDetail", r.URL.Path)
		}
		if r.URL.Query().Get("plz") != "800100" {
			t.Errorf("plz = %q, want 800100", r.URL.Query().Get("plz"))
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := &Client{http: &http.Client{}, baseURL: server.URL, lang: "en"}
	detail, err := c.GetForecast("800100")
	if err != nil {
		t.Fatalf("GetForecast error = %v", err)
	}
	if detail.CurrentWeather.Temperature != 16.8 {
		t.Errorf("temperature = %f, want 16.8", detail.CurrentWeather.Temperature)
	}
	if len(detail.Forecast) != 2 {
		t.Errorf("forecast days = %d, want 2", len(detail.Forecast))
	}
}
