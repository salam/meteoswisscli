package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// INCAResponse represents the JSON structure of an INCA precipitation frame.
type INCAResponse struct {
	Coords INCACoords `json:"coords"`
	Areas  []INCAArea `json:"areas"`
}

// INCACoords describes the grid coordinate system.
type INCACoords struct {
	System string  `json:"system"`
	XMin   float64 `json:"x_min"`
	XMax   float64 `json:"x_max"`
	XCount int     `json:"x_count"`
	YMin   float64 `json:"y_min"`
	YMax   float64 `json:"y_max"`
	YCount int     `json:"y_count"`
}

// INCAArea represents a colored precipitation area with shapes.
type INCAArea struct {
	Color  string        `json:"color"`
	Shapes [][]INCAShape `json:"shapes"`
}

// INCAShape represents a single vector-encoded polygon in the INCA grid.
type INCAShape struct {
	I int    `json:"i"`
	J int    `json:"j"`
	D string `json:"d"`
	O string `json:"o"`
	L int    `json:"l"`
}

// INCAFrame represents a rendered INCA precipitation frame as a simple grid.
type INCAFrame struct {
	Timestamp string
	Rows      int
	Cols      int
	Data      []float64 // row-major precipitation intensity (0-10 scale)
}

const incaBaseURL = "https://www.meteoschweiz.admin.ch/product/output"

// GetINCAVersion fetches the current INCA precipitation rate version.
func (c *Client) GetINCAVersion() (string, error) {
	data, err := c.DoRaw("GET", incaBaseURL+"/versions.json")
	if err != nil {
		return "", fmt.Errorf("fetch versions: %w", err)
	}
	var versions map[string]string
	if err := json.Unmarshal(data, &versions); err != nil {
		return "", fmt.Errorf("parse versions: %w", err)
	}
	ver, ok := versions["inca/precipitation/rate"]
	if !ok {
		return "", fmt.Errorf("INCA precipitation version not found")
	}
	return ver, nil
}

// GetINCAFrame fetches and rasterizes an INCA precipitation frame.
func (c *Client) GetINCAFrame(version, timestamp string) (*INCAFrame, error) {
	url := fmt.Sprintf("%s/inca/precipitation/rate/version__%s/rate_%s.json", incaBaseURL, version, timestamp)
	data, err := c.DoRaw("GET", url)
	if err != nil {
		return nil, fmt.Errorf("fetch INCA frame: %w", err)
	}

	var resp INCAResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse INCA frame: %w", err)
	}

	return rasterizeINCA(&resp, timestamp), nil
}

// ListINCATimestamps returns available 5-minute timestamps for a given INCA version.
// Instead of probing every timestamp (slow), it generates expected timestamps based
// on the version time. INCA typically provides frames from ~version-time to ~version+6h.
func (c *Client) ListINCATimestamps(version string, count int) ([]string, error) {
	verTime, err := time.Parse("20060102_1504", version)
	if err != nil {
		return nil, fmt.Errorf("parse version time: %w", err)
	}

	// Round version time up to next 5-minute boundary for frame start
	start := verTime.Truncate(5 * time.Minute)
	if start.Before(verTime) {
		start = start.Add(5 * time.Minute)
	}
	end := start.Add(6 * time.Hour)

	// Quick probe to confirm the start exists; if not, try next slot
	for probe := start; probe.Before(start.Add(30 * time.Minute)); probe = probe.Add(5 * time.Minute) {
		ts := probe.Format("20060102_1504")
		url := fmt.Sprintf("%s/inca/precipitation/rate/version__%s/rate_%s.json", incaBaseURL, version, ts)
		if resp, err := c.http.Head(url); err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				start = probe
				break
			}
		}
	}

	// Find the actual end by probing backwards from +6h
	found := false
	for probe := end; probe.After(start); probe = probe.Add(-15 * time.Minute) {
		ts := probe.Format("20060102_1504")
		url := fmt.Sprintf("%s/inca/precipitation/rate/version__%s/rate_%s.json", incaBaseURL, version, ts)
		resp, err := c.http.Head(url)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 200 {
			end = probe
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("no INCA frames found for version %s", version)
	}

	// Generate all 5-minute timestamps in the range
	var timestamps []string
	for t := start; !t.After(end); t = t.Add(5 * time.Minute) {
		timestamps = append(timestamps, t.Format("20060102_1504"))
	}

	// Limit to last N timestamps if requested
	if count > 0 && len(timestamps) > count {
		timestamps = timestamps[len(timestamps)-count:]
	}

	return timestamps, nil
}

// colorIntensity maps INCA hex colors to precipitation intensity (mm/h).
var colorIntensity = map[string]float64{
	"9a7e95": -1,   // border (skip)
	"333e48": -2,   // lakes (skip)
	"cccccc": 0.05, // trace
	"04fdff": 0.1,  // very light
	"0001fc": 5.0,  // heavy rain
	"058c2d": 2.0,  // moderate rain
	"05ff05": 1.0,  // light-moderate rain
	"feff01": 0.5,  // light rain
	"ffc703": 3.0,  // moderate-heavy rain
	"ff0100": 10.0, // extreme
	"fe0000": 10.0, // extreme (variant)
}

func rasterizeINCA(resp *INCAResponse, timestamp string) *INCAFrame {
	rows := resp.Coords.YCount
	cols := resp.Coords.XCount
	if rows <= 0 {
		rows = 640
	}
	if cols <= 0 {
		cols = 710
	}

	grid := make([]float64, rows*cols)

	for _, area := range resp.Areas {
		intensity, ok := colorIntensity[area.Color]
		if !ok || intensity < 0 {
			continue // skip borders, lakes, unknown
		}

		for _, shapeGroup := range area.Shapes {
			for _, shape := range shapeGroup {
				rasterizeShape(grid, rows, cols, shape, intensity)
			}
		}
	}

	return &INCAFrame{
		Timestamp: timestamp,
		Rows:      rows,
		Cols:      cols,
		Data:      grid,
	}
}

// rasterizeShape decodes the direction+opacity encoding and fills pixels.
// The encoding works like a scanline fill:
// - 'd' traces the polygon boundary: N=down, L=up, M=right, K=left, O=down-right
// - 'o' encodes fill widths at scanline crossings (hex digits)
func rasterizeShape(grid []float64, rows, cols int, shape INCAShape, intensity float64) {
	y := shape.I
	x := shape.J
	oIdx := 0

	for _, ch := range shape.D {
		// Get fill width from 'o' encoding
		fillWidth := 0
		if oIdx < len(shape.O) {
			c := shape.O[oIdx]
			if c >= '0' && c <= '9' {
				fillWidth = int(c - '0')
			} else if c >= 'a' && c <= 'f' {
				fillWidth = int(c-'a') + 10
			}
			oIdx++
		}

		// Fill pixels
		for dx := 0; dx < fillWidth; dx++ {
			px := x + dx
			if y >= 0 && y < rows && px >= 0 && px < cols {
				grid[y*cols+px] = intensity
			}
		}

		// Move according to direction
		switch ch {
		case 'N':
			y++
		case 'L':
			y--
		case 'M':
			x++
		case 'K':
			x--
		case 'O':
			y++
			x++
		}
	}
}
