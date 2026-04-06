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

// Satellite/cloud image URLs (latest snapshots from MeteoSwiss open data).
var satelliteImageURLs = map[RadarType]string{
	RadarSatellite: "https://www.meteoschweiz.admin.ch/static/resources/satellite/satellite-hrv-latest.png",
	RadarCloud:     "https://www.meteoschweiz.admin.ch/static/resources/satellite/satellite-infrared-latest.png",
}

// GetSatelliteImageURL returns the latest satellite/cloud image URL for the given type.
func GetSatelliteImageURL(rt RadarType) string { return satelliteImageURLs[rt] }

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

// ListRadarFrames returns available radar frames, probing beyond the STAC index
// to find the most recent data (STAC lags ~2h, but files are accessible earlier).
func (c *Client) ListRadarFrames(limit int) ([]RadarFrame, error) {
	now := time.Now().UTC()
	today := now.Format("20060102")
	yy := now.Format("06")
	doy := fmt.Sprintf("%03d", now.YearDay())

	// Step 1: Get frames from STAC index
	stacURL := fmt.Sprintf("%s/collections/ch.meteoschweiz.ogd-radar-precip/items/%s-ch", stacBaseURL, today)
	data, err := c.DoRaw("GET", stacURL)
	if err != nil {
		yesterday := now.AddDate(0, 0, -1)
		yy = yesterday.Format("06")
		doy = fmt.Sprintf("%03d", yesterday.YearDay())
		today = yesterday.Format("20060102")
		stacURL = fmt.Sprintf("%s/collections/ch.meteoschweiz.ogd-radar-precip/items/%s-ch", stacBaseURL, today)
		data, err = c.DoRaw("GET", stacURL)
		if err != nil {
			return nil, fmt.Errorf("fetch radar items: %w", err)
		}
	}

	var feature struct {
		ID     string `json:"id"`
		Assets map[string]struct {
			Href string `json:"href"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(data, &feature); err != nil {
		return nil, fmt.Errorf("parse STAC response: %w", err)
	}

	// Collect CPC frames from STAC
	var frames []RadarFrame
	var latestTime time.Time
	for name, asset := range feature.Assets {
		if t, ok := parseCPCFilename(name); ok {
			frames = append(frames, RadarFrame{Timestamp: t.Format("2006-01-02 15:04"), URL: asset.Href, Rows: 640, Cols: 710})
			if t.After(latestTime) {
				latestTime = t
			}
		}
	}

	// Step 2: Probe forward from latest STAC entry to find newer files
	// Files are published ~5 min apart, public access lags ~2h but files exist earlier
	if !latestTime.IsZero() {
		baseURL := fmt.Sprintf("%s/ch.meteoschweiz.ogd-radar-precip/%s-ch", "https://data.geo.admin.ch", today)
		probeTime := latestTime.Add(5 * time.Minute)
		for probeTime.Before(now) {
			fname := fmt.Sprintf("cpc%s%s%s0_00060.001.h5", yy, doy, probeTime.Format("1504"))
			probeURL := baseURL + "/" + fname
			resp, err := http.Head(probeURL)
			if err != nil || resp.StatusCode != 200 {
				break
			}
			resp.Body.Close()
			frames = append(frames, RadarFrame{
				Timestamp: probeTime.Format("2006-01-02 15:04"),
				URL:       probeURL,
				Rows:      640, Cols: 710,
			})
			probeTime = probeTime.Add(5 * time.Minute)
		}
	}

	sort.Slice(frames, func(i, j int) bool { return frames[i].Timestamp < frames[j].Timestamp })

	if limit > 0 && len(frames) > limit {
		frames = frames[len(frames)-limit:]
	}

	return frames, nil
}

// parseCPCFilename parses a CPC HDF5 filename into a timestamp.
// Format: cpcYYDDDHHMM0_00060.001.h5 (YY=year, DDD=day-of-year, HHMM=time)
func parseCPCFilename(name string) (time.Time, bool) {
	if !strings.HasPrefix(name, "cpc") || !strings.HasSuffix(name, ".h5") {
		return time.Time{}, false
	}
	digits := name[3:]
	if len(digits) < 10 {
		return time.Time{}, false
	}
	year := 2000 + atoi(digits[0:2])
	doyVal := atoi(digits[2:5])
	hour := atoi(digits[5:7])
	minute := atoi(digits[7:9])
	if doyVal < 1 || doyVal > 366 || hour > 23 || minute > 59 {
		return time.Time{}, false
	}
	t := time.Date(year, 1, 1, hour, minute, 0, 0, time.UTC).AddDate(0, 0, doyVal-1)
	return t, true
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
