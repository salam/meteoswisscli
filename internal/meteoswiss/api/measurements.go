package api

import (
	"encoding/csv"
	"fmt"
	"strings"
)

const measurementsURL = "https://data.geo.admin.ch/ch.meteoschweiz.messwerte-aktuell/VQHA80.csv"

type StationMeasurement struct {
	Station     string `json:"station"`
	Date        string `json:"date"`
	Temperature string `json:"temperature"`
	Rainfall    string `json:"rainfall"`
	Sunshine    string `json:"sunshine"`
	Radiation   string `json:"radiation"`
	Humidity    string `json:"humidity"`
	DewPoint    string `json:"dew_point"`
	WindDir     string `json:"wind_direction"`
	WindSpeed   string `json:"wind_speed"`
	GustPeak    string `json:"gust_peak"`
	Pressure    string `json:"pressure_station"`
	PressureQFE string `json:"pressure_qfe"`
	PressureQNH string `json:"pressure_qnh"`
}

func (c *Client) GetCurrentMeasurements(url string) ([]StationMeasurement, error) {
	if url == "" {
		url = measurementsURL
	}

	data, err := c.DoRaw("GET", url)
	if err != nil {
		return nil, fmt.Errorf("fetch measurements: %w", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = ';'
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("no measurement data found")
	}

	var measurements []StationMeasurement
	for _, row := range records[1:] {
		if len(row) < 14 {
			continue
		}
		m := StationMeasurement{
			Station:     row[0],
			Date:        row[1],
			Temperature: dashToEmpty(row[2]),
			Rainfall:    dashToEmpty(row[3]),
			Sunshine:    dashToEmpty(row[4]),
			Radiation:   dashToEmpty(row[5]),
			Humidity:    dashToEmpty(row[6]),
			DewPoint:    dashToEmpty(row[7]),
			WindDir:     dashToEmpty(row[8]),
			WindSpeed:   dashToEmpty(row[9]),
			GustPeak:    dashToEmpty(row[10]),
			Pressure:    dashToEmpty(row[11]),
			PressureQFE: dashToEmpty(row[12]),
			PressureQNH: dashToEmpty(row[13]),
		}
		measurements = append(measurements, m)
	}
	return measurements, nil
}

func dashToEmpty(s string) string {
	if s == "-" {
		return ""
	}
	return s
}
