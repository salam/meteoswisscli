package api

import "fmt"

type IMISStation struct {
	Code        string  `json:"code"`
	Label       string  `json:"label"`
	Lon         float64 `json:"lon"`
	Lat         float64 `json:"lat"`
	Elevation   float64 `json:"elevation"`
	CountryCode string  `json:"country_code"`
	CantonCode  string  `json:"canton_code"`
	Type        string  `json:"type"`
}

type StudyPlotStation struct {
	Code        string  `json:"code"`
	Label       string  `json:"label"`
	Lon         float64 `json:"lon"`
	Lat         float64 `json:"lat"`
	Elevation   float64 `json:"elevation"`
	CountryCode string  `json:"country_code"`
	CantonCode  string  `json:"canton_code"`
}

type IMISMeasurement struct {
	StationCode   string   `json:"station_code"`
	MeasureDate   string   `json:"measure_date"`
	HS            *float64 `json:"HS"`
	TA30MinMean   *float64 `json:"TA_30MIN_MEAN"`
	RH30MinMean   *float64 `json:"RH_30MIN_MEAN"`
	TSS30MinMean  *float64 `json:"TSS_30MIN_MEAN"`
	VW30MinMean   *float64 `json:"VW_30MIN_MEAN"`
	VW30MinMax    *float64 `json:"VW_30MIN_MAX"`
	DW30MinMean   *float64 `json:"DW_30MIN_MEAN"`
	RSWR30MinMean *float64 `json:"RSWR_30MIN_MEAN"`
}

type StudyPlotMeasurement struct {
	StationCode string   `json:"station_code"`
	MeasureDate string   `json:"measure_date"`
	HS          *float64 `json:"HS"`
	HN1D        *float64 `json:"HN_1D"`
	HNW1D       *float64 `json:"HNW_1D"`
}

func (c *Client) GetIMISStations() ([]IMISStation, error) {
	url := fmt.Sprintf("%s/public/api/imis/stations", c.measurementBase)
	var result []IMISStation
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get IMIS stations: %w", err)
	}
	return result, nil
}

func (c *Client) GetStudyPlotStations() ([]StudyPlotStation, error) {
	url := fmt.Sprintf("%s/public/api/study-plot/stations", c.measurementBase)
	var result []StudyPlotStation
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get study plot stations: %w", err)
	}
	return result, nil
}

func (c *Client) GetIMISMeasurementsByStation(code string, periodDays int) ([]IMISMeasurement, error) {
	url := fmt.Sprintf("%s/public/api/imis/station/%s/measurements?period_in_days=%d", c.measurementBase, code, periodDays)
	var result []IMISMeasurement
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get IMIS measurements: %w", err)
	}
	return result, nil
}

func (c *Client) GetStudyPlotMeasurementsByStation(code string, periodDays int) ([]StudyPlotMeasurement, error) {
	url := fmt.Sprintf("%s/public/api/study-plot/station/%s/measurements?period_in_days=%d", c.measurementBase, code, periodDays)
	var result []StudyPlotMeasurement
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get study plot measurements: %w", err)
	}
	return result, nil
}
