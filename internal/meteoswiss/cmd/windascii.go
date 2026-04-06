package cmd

import (
	"strconv"
	"strings"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/geo"
	"github.com/salam/swissmeteocli/pkg/output"
)

func renderWindASCII(measurements []api.StationMeasurement, width int, noColor bool) string {
	const chMinLat = 45.8
	const chMaxLat = 47.85
	const chMinLon = 5.9
	const chMaxLon = 10.55

	height := width * 2 / 5 // approximate aspect ratio for Switzerland

	// Create character grid
	grid := make([][]string, height)
	for y := range grid {
		grid[y] = make([]string, width)
		for x := range grid[y] {
			grid[y][x] = " "
		}
	}

	// Overlay Swiss border and lakes
	borderGrid, lakeGrid := output.RenderOverlay(width, height, chMinLat, chMaxLat, chMinLon, chMaxLon)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if borderGrid[y][x] {
				if noColor {
					grid[y][x] = "·"
				} else {
					grid[y][x] = "\033[38;5;240m·\033[0m"
				}
			}
			if lakeGrid[y][x] {
				if noColor {
					grid[y][x] = "~"
				} else {
					grid[y][x] = "\033[38;5;33m~\033[0m"
				}
			}
		}
	}

	// Place wind arrows for each station
	for _, m := range measurements {
		if m.WindSpeed == "" || m.WindDir == "" {
			continue
		}
		station := geo.LookupStation(m.Station)
		if station == nil {
			continue
		}

		// Convert lat/lon to grid position
		x := int((station.Lon - chMinLon) / (chMaxLon - chMinLon) * float64(width))
		y := int((chMaxLat - station.Lat) / (chMaxLat - chMinLat) * float64(height))

		if x < 0 || x >= width || y < 0 || y >= height {
			continue
		}

		dir, _ := strconv.ParseFloat(m.WindDir, 64)
		speed, _ := strconv.ParseFloat(m.WindSpeed, 64)

		arrow := directionArrow(dir)
		colored := colorBySpeed(arrow, speed, noColor)
		grid[y][x] = colored
	}

	// Render grid to string
	var sb strings.Builder
	for _, row := range grid {
		for _, cell := range row {
			sb.WriteString(cell)
		}
		sb.WriteString("\n")
	}

	// Legend
	if noColor {
		sb.WriteString("\nWind: \u2191\u2197\u2192\u2198\u2193\u2199\u2190\u2196 direction  \u00b7 border  ~ lake\n")
	} else {
		sb.WriteString("\nWind: \u2191\u2197\u2192\u2198\u2193\u2199\u2190\u2196 direction  ")
		sb.WriteString("\033[37m\u25cb\033[0m <10  ")
		sb.WriteString("\033[36m\u25cb\033[0m <20  ")
		sb.WriteString("\033[32m\u25cb\033[0m <40  ")
		sb.WriteString("\033[33m\u25cb\033[0m <60  ")
		sb.WriteString("\033[38;5;208m\u25cb\033[0m <80  ")
		sb.WriteString("\033[31m\u25cb\033[0m >80 km/h  ")
		sb.WriteString("\033[38;5;240m\u00b7\033[0m border  ")
		sb.WriteString("\033[38;5;33m~\033[0m lake\n")
	}

	return sb.String()
}

func directionArrow(degrees float64) string {
	for degrees < 0 {
		degrees += 360
	}
	for degrees >= 360 {
		degrees -= 360
	}

	switch {
	case degrees < 22.5 || degrees >= 337.5:
		return "\u2191" // N
	case degrees < 67.5:
		return "\u2197" // NE
	case degrees < 112.5:
		return "\u2192" // E
	case degrees < 157.5:
		return "\u2198" // SE
	case degrees < 202.5:
		return "\u2193" // S
	case degrees < 247.5:
		return "\u2199" // SW
	case degrees < 292.5:
		return "\u2190" // W
	default:
		return "\u2196" // NW
	}
}

func colorBySpeed(arrow string, speed float64, noColor bool) string {
	if noColor {
		return arrow
	}
	switch {
	case speed < 10:
		return "\033[37m" + arrow + "\033[0m" // white
	case speed < 20:
		return "\033[36m" + arrow + "\033[0m" // cyan
	case speed < 40:
		return "\033[32m" + arrow + "\033[0m" // green
	case speed < 60:
		return "\033[33m" + arrow + "\033[0m" // yellow
	case speed < 80:
		return "\033[38;5;208m" + arrow + "\033[0m" // orange
	default:
		return "\033[31m" + arrow + "\033[0m" // red
	}
}
