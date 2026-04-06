package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type RadarType string

const (
	RadarRain      RadarType = "rain"
	RadarCloud     RadarType = "cloud"
	RadarSatellite RadarType = "satellite"
)

var radarBrowserURLs = map[RadarType]string{
	RadarRain:      "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/niederschlag.html",
	RadarCloud:     "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/satellitenbilder.html#tab=satellite-animation-hrv",
	RadarSatellite: "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/satellitenbilder.html#tab=satellite-animation-hrv",
}

func GetRadarBrowserURL(rt RadarType) string { return radarBrowserURLs[rt] }

// RadarFrame represents a single radar precipitation snapshot.
type RadarFrame struct {
	Timestamp string     `json:"timestamp"`
	URL       string     `json:"url"`
	Rows      int        `json:"rows"`
	Cols      int        `json:"cols"`
	Data      [][]float64 `json:"-"` // precipitation in mm, loaded on demand
}

const stacBaseURL = "https://data.geo.admin.ch/api/stac/v1"
const radarDataBaseURL = "https://data.geo.admin.ch/ch.meteoschweiz.ogd-radar-precip"

// ListRadarFrames returns available radar frames from the STAC API for today.
// Returns CPC (CombiPrecip) HDF5 asset URLs sorted by timestamp.
func (c *Client) ListRadarFrames(limit int) ([]RadarFrame, error) {
	// Construct today's item ID: YYYYMMDD-ch
	today := time.Now().UTC().Format("20060102")
	url := fmt.Sprintf("%s/collections/ch.meteoschweiz.ogd-radar-precip/items/%s-ch", stacBaseURL, today)

	data, err := c.DoRaw("GET", url)
	if err != nil {
		// Try yesterday if today's not available yet
		yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("20060102")
		url = fmt.Sprintf("%s/collections/ch.meteoschweiz.ogd-radar-precip/items/%s-ch", stacBaseURL, yesterday)
		data, err = c.DoRaw("GET", url)
		if err != nil {
			return nil, fmt.Errorf("fetch radar items: %w", err)
		}
	}

	var feature struct {
		ID     string `json:"id"`
		Assets map[string]struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(data, &feature); err != nil {
		return nil, fmt.Errorf("parse STAC response: %w", err)
	}

	// Extract CPC (CombiPrecip) assets
	// Filename format: cpcYYDDDHHMM0_00060.001.h5
	// YY=year, DDD=day-of-year, HHMM=hour:minute, trailing 0
	var frames []RadarFrame
	for name, asset := range feature.Assets {
		if !strings.HasPrefix(name, "cpc") || !strings.HasSuffix(name, ".h5") {
			continue
		}
		// Parse: cpc + YYDDD + HHMM + 0 + _00060.001.h5
		digits := name[3:] // strip "cpc"
		if len(digits) < 10 {
			continue
		}
		yy := digits[0:2]
		ddd := digits[2:5]
		hh := digits[5:7]
		mm := digits[7:9]

		year := 2000 + atoi(yy)
		doy := atoi(ddd)
		hour := atoi(hh)
		minute := atoi(mm)

		if doy < 1 || doy > 366 || hour > 23 || minute > 59 {
			continue
		}

		t := time.Date(year, 1, 1, hour, minute, 0, 0, time.UTC).AddDate(0, 0, doy-1)
		frames = append(frames, RadarFrame{
			Timestamp: t.Format("2006-01-02 15:04"),
			URL:       asset.Href,
			Rows:      640,
			Cols:      710,
		})
	}

	sort.Slice(frames, func(i, j int) bool { return frames[i].Timestamp < frames[j].Timestamp })

	// Return last N frames
	if limit > 0 && len(frames) > limit {
		frames = frames[len(frames)-limit:]
	}

	return frames, nil
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// DownloadRadarH5 downloads a raw HDF5 file.
func (c *Client) DownloadRadarH5(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download radar data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("radar download error %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
