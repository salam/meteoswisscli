package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCurrentWeather_TimeFormatted(t *testing.T) {
	cw := CurrentWeather{
		Time:        1712332800000, // 2024-04-05 16:00 UTC
		Temperature: 18.5,
	}
	formatted := cw.TimeFormatted()
	if formatted == "" {
		t.Error("TimeFormatted() should not be empty")
	}
	// Should contain a date pattern
	if len(formatted) < 10 {
		t.Errorf("TimeFormatted() = %q, expected date-time string", formatted)
	}
}

func TestCurrentWeather_TimeFormatted_Zero(t *testing.T) {
	cw := CurrentWeather{Time: 0}
	formatted := cw.TimeFormatted()
	if formatted == "" {
		t.Error("TimeFormatted() should not be empty even for zero time")
	}
}

func TestWarnTypeName_Known(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "Wind"},
		{1, "Thunderstorm"},
		{2, "Rain"},
		{3, "Snow"},
		{4, "Slippery roads"},
		{5, "Frost"},
		{6, "Heat wave"},
		{7, "Forest fire"},
		{8, "Avalanche"},
		{9, "Earthquake"},
		{10, "Flood"},
	}
	for _, tt := range tests {
		got := WarnTypeName(tt.input)
		if got != tt.want {
			t.Errorf("WarnTypeName(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWarnTypeName_Unknown(t *testing.T) {
	got := WarnTypeName(99)
	if got != "Type 99" {
		t.Errorf("WarnTypeName(99) = %q, want %q", got, "Type 99")
	}
}

func TestWarning_ValidFromFormatted(t *testing.T) {
	w := Warning{ValidFrom: 1712332800000}
	f := w.ValidFromFormatted()
	if f == "" || len(f) < 10 {
		t.Errorf("ValidFromFormatted() = %q, expected date-time string", f)
	}
}

func TestWarning_ValidToFormatted(t *testing.T) {
	w := Warning{ValidTo: 1712419200000}
	f := w.ValidToFormatted()
	if f == "" || len(f) < 10 {
		t.Errorf("ValidToFormatted() = %q, expected date-time string", f)
	}
}

func TestIconDescription_Known(t *testing.T) {
	tests := []struct {
		id   int
		want string
	}{
		{1, "Sunny"},
		{2, "Mostly sunny"},
		{7, "Rain"},
		{101, "Clear night"},
		{109, "Thunderstorm night"},
	}
	for _, tt := range tests {
		got := IconDescription(tt.id)
		if got != tt.want {
			t.Errorf("IconDescription(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestIconDescription_Unknown(t *testing.T) {
	got := IconDescription(999)
	if got != "Icon 999" {
		t.Errorf("IconDescription(999) = %q, want %q", got, "Icon 999")
	}
}

func TestWeatherIconURL(t *testing.T) {
	url := WeatherIconURL(1)
	if url != "https://www.meteoschweiz.admin.ch/static/resources/weather-symbols/1.svg" {
		t.Errorf("WeatherIconURL(1) = %q", url)
	}
}

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
