package api

import (
	"encoding/csv"
	"fmt"
	"strings"
)

type PollenStation struct {
	Code string
	Name string
	Lat  float64
	Lon  float64
}

var PollenStations = []PollenStation{
	{"PBE", "Bern", 46.9503, 7.4247},
	{"PBS", "Basel", 47.5618, 7.5839},
	{"PBU", "Buchs SG", 47.1733, 9.4726},
	{"PCF", "La Chaux-de-Fonds", 47.1135, 6.832},
	{"PDS", "Davos", 46.8291, 9.8555},
	{"PGE", "Genève", 46.192, 6.1475},
	{"PJU", "Jungfraujoch", 46.5475, 7.985},
	{"PLO", "Locarno", 46.1725, 8.7874},
	{"PLS", "Lausanne", 46.5241, 6.6448},
	{"PLU", "Lugano", 46.0042, 8.9606},
	{"PLZ", "Luzern", 47.0577, 8.2968},
	{"PMU", "Münsterlingen", 47.6302, 9.2369},
	{"PNE", "Neuchâtel", 47.0003, 6.9498},
	{"PPY", "Payerne", 46.8134, 6.9429},
	{"PSN", "Sion", 46.2354, 7.3846},
	{"PZH", "Zürich", 47.3782, 8.5656},
}

type PollenMeasurement struct {
	Station string `json:"station"`
	Date    string `json:"date"`
	Alder   string `json:"alder"`
	Birch   string `json:"birch"`
	Hazel   string `json:"hazel"`
	Beech   string `json:"beech"`
	Ash     string `json:"ash"`
	Oak     string `json:"oak"`
	Grasses string `json:"grasses"`
}

func (c *Client) GetPollenData(stationCode string) ([]PollenMeasurement, error) {
	url := fmt.Sprintf("https://data.geo.admin.ch/ch.meteoschweiz.ogd-pollen/%s/ogd-pollen_%s_d_recent.csv",
		strings.ToLower(stationCode), strings.ToLower(stationCode))

	data, err := c.DoRaw("GET", url)
	if err != nil {
		return nil, fmt.Errorf("fetch pollen data: %w", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = ';'
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse pollen CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("no pollen data found")
	}

	var measurements []PollenMeasurement
	for _, row := range records[1:] {
		if len(row) < 8 {
			continue
		}
		measurements = append(measurements, PollenMeasurement{
			Station: row[0],
			Date:    row[1],
			Alder:   row[2],
			Birch:   row[3],
			Hazel:   row[4],
			Beech:   row[5],
			Ash:     row[6],
			Oak:     row[7],
			Grasses: row[8],
		})
	}
	return measurements, nil
}
