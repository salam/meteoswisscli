package api

import (
	"fmt"
	"time"
)

type PlzDetail struct {
	CurrentWeather CurrentWeather `json:"currentWeather"`
	Forecast       []ForecastDay  `json:"forecast"`
	Warnings       []Warning      `json:"warnings,omitempty"`
	Graph          *Graph         `json:"graph,omitempty"`
}

type CurrentWeather struct {
	Time        int64   `json:"time"`
	Icon        int     `json:"icon"`
	IconV2      int     `json:"iconV2"`
	Temperature float64 `json:"temperature"`
}

func (cw CurrentWeather) TimeFormatted() string {
	return time.UnixMilli(cw.Time).Format("2006-01-02 15:04")
}

type ForecastDay struct {
	DayDate          string  `json:"dayDate"`
	IconDay          int     `json:"iconDay"`
	IconDayV2        int     `json:"iconDayV2"`
	TemperatureMax   float64 `json:"temperatureMax"`
	TemperatureMin   float64 `json:"temperatureMin"`
	Precipitation    float64 `json:"precipitation"`
	PrecipitationMin float64 `json:"precipitationMin"`
	PrecipitationMax float64 `json:"precipitationMax"`
}

type Warning struct {
	Type      int           `json:"warnType"`
	Level     int           `json:"warnLevel"`
	ValidFrom int64         `json:"validFrom"`
	ValidTo   int64         `json:"validTo"`
	Ordering  string        `json:"ordering,omitempty"`
	Outlook   bool          `json:"outlook,omitempty"`
	Links     []WarningLink `json:"links,omitempty"`
}

type WarningLink struct {
	URL    string `json:"url"`
	Text   string `json:"text"`
	AltURL string `json:"altUrl,omitempty"`
}

func (w Warning) ValidFromFormatted() string {
	return time.UnixMilli(w.ValidFrom).Format("2006-01-02 15:04")
}

func (w Warning) ValidToFormatted() string {
	return time.UnixMilli(w.ValidTo).Format("2006-01-02 15:04")
}

var warnTypeNames = map[int]string{
	0: "Wind", 1: "Thunderstorm", 2: "Rain", 3: "Snow", 4: "Slippery roads",
	5: "Frost", 6: "Heat wave", 7: "Forest fire", 8: "Avalanche",
	9: "Earthquake", 10: "Flood",
}

func WarnTypeName(t int) string {
	if name, ok := warnTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Type %d", t)
}

type Graph struct {
	Start                      int64     `json:"start"`
	StartLowResolution         int64     `json:"startLowResolution"`
	Precipitation10m           []float64 `json:"precipitation10m,omitempty"`
	PrecipitationMin10m        []float64 `json:"precipitationMin10m,omitempty"`
	PrecipitationMax10m        []float64 `json:"precipitationMax10m,omitempty"`
	Precipitation1h            []float64 `json:"precipitation1h,omitempty"`
	PrecipitationMin1h         []float64 `json:"precipitationMin1h,omitempty"`
	PrecipitationMax1h         []float64 `json:"precipitationMax1h,omitempty"`
	PrecipitationProbability3h []float64 `json:"precipitationProbability3h,omitempty"`
	TemperatureMean1h          []float64 `json:"temperatureMean1h,omitempty"`
	TemperatureMin1h           []float64 `json:"temperatureMin1h,omitempty"`
	TemperatureMax1h           []float64 `json:"temperatureMax1h,omitempty"`
	WindSpeed1h                []float64 `json:"windSpeed1h,omitempty"`
	WindGust1h                 []float64 `json:"gustSpeed1h,omitempty"`
	WindDirection3h            []int     `json:"windDirection3h,omitempty"`
	Sunshine1h                 []int     `json:"sunshine1h,omitempty"`
	WeatherIcon3h              []int     `json:"weatherIcon3h,omitempty"`
	Sunrise                    []int64   `json:"sunrise,omitempty"`
	Sunset                     []int64   `json:"sunset,omitempty"`
}

func (c *Client) GetForecast(plz string) (*PlzDetail, error) {
	var detail PlzDetail
	err := c.DoJSON("GET", fmt.Sprintf("/v1/plzDetail?plz=%s", plz), nil, &detail)
	if err != nil {
		return nil, fmt.Errorf("get forecast: %w", err)
	}
	return &detail, nil
}
