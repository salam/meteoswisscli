package api

import "fmt"

type PlzDetail struct {
	CurrentWeather CurrentWeather `json:"currentWeather"`
	Forecast       []ForecastDay  `json:"forecast"`
	Warnings       []Warning      `json:"warnings,omitempty"`
	Graph          *Graph         `json:"graph,omitempty"`
}

type CurrentWeather struct {
	Time        string  `json:"time"`
	Icon        int     `json:"icon"`
	IconV2      int     `json:"iconV2"`
	Temperature float64 `json:"temperature"`
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
	Type      int    `json:"warnType"`
	Level     int    `json:"warnLevel"`
	Text      string `json:"text"`
	ValidFrom string `json:"validFrom"`
	ValidTo   string `json:"validTo"`
}

type Graph struct {
	Start              string    `json:"start"`
	StartLowResolution string    `json:"startLowResolution"`
	Precipitation10m   []float64 `json:"precipitation10m,omitempty"`
	Precipitation1h    []float64 `json:"precipitation1h,omitempty"`
	TemperatureMean1h  []float64 `json:"temperatureMean1h,omitempty"`
	TemperatureMin1h   []float64 `json:"temperatureMin1h,omitempty"`
	TemperatureMax1h   []float64 `json:"temperatureMax1h,omitempty"`
	WindSpeed1h        []float64 `json:"windSpeed1h,omitempty"`
	WindGust1h         []float64 `json:"windGust1h,omitempty"`
	WindDirection1h    []int     `json:"windDirection1h,omitempty"`
	Sunrise            []string  `json:"sunrise,omitempty"`
	Sunset             []string  `json:"sunset,omitempty"`
}

func (c *Client) GetForecast(plz string) (*PlzDetail, error) {
	var detail PlzDetail
	err := c.DoJSON("GET", fmt.Sprintf("/v1/plzDetail?plz=%s", plz), nil, &detail)
	if err != nil {
		return nil, fmt.Errorf("get forecast: %w", err)
	}
	return &detail, nil
}
