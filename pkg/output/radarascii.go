package output

import (
	"strings"
)

// RadarGrid holds a 2D precipitation grid extracted from an HDF5 file.
type RadarGrid struct {
	Rows   int
	Cols   int
	Data   []float64 // row-major, NaN = no data
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// precipToRune maps precipitation intensity (mm/h) to a colored Unicode character.
func precipToRune(val float64) string {
	if val <= 0 {
		return " "
	}
	// Precipitation intensity scale (mm/h for 1h accumulation)
	switch {
	case val < 0.1:
		return "\033[38;5;39m░\033[0m" // light blue - trace
	case val < 0.5:
		return "\033[38;5;33m▒\033[0m" // blue - light rain
	case val < 1.0:
		return "\033[38;5;28m▓\033[0m" // green - moderate
	case val < 2.0:
		return "\033[38;5;226m▓\033[0m" // yellow - heavy
	case val < 5.0:
		return "\033[38;5;208m█\033[0m" // orange - very heavy
	default:
		return "\033[38;5;196m█\033[0m" // red - extreme
	}
}

func precipToRuneNoColor(val float64) string {
	if val <= 0 {
		return " "
	}
	switch {
	case val < 0.1:
		return "░"
	case val < 0.5:
		return "▒"
	case val < 2.0:
		return "▓"
	default:
		return "█"
	}
}

// RenderRadarASCII renders a radar grid as colored ASCII art.
// showBorder and showLakes control whether the Swiss border and lakes are overlaid.
func RenderRadarASCII(grid *RadarGrid, width int, noColor bool, showBorder, showLakes bool) string {
	if grid.Rows == 0 || grid.Cols == 0 {
		return ""
	}

	// Crop to Switzerland (approximate lat/lon bounds)
	// Switzerland: roughly 45.8-47.8 N, 5.9-10.5 E
	const chMinLat = 45.8
	const chMaxLat = 47.85
	const chMinLon = 5.9
	const chMaxLon = 10.55

	// Map lat/lon to grid indices
	latRange := grid.MaxLat - grid.MinLat
	lonRange := grid.MaxLon - grid.MinLon

	startRow := int(float64(grid.Rows) * (grid.MaxLat - chMaxLat) / latRange)
	endRow := int(float64(grid.Rows) * (grid.MaxLat - chMinLat) / latRange)
	startCol := int(float64(grid.Cols) * (chMinLon - grid.MinLon) / lonRange)
	endCol := int(float64(grid.Cols) * (chMaxLon - grid.MinLon) / lonRange)

	if startRow < 0 {
		startRow = 0
	}
	if endRow > grid.Rows {
		endRow = grid.Rows
	}
	if startCol < 0 {
		startCol = 0
	}
	if endCol > grid.Cols {
		endCol = grid.Cols
	}

	cropRows := endRow - startRow
	cropCols := endCol - startCol

	if cropRows <= 0 || cropCols <= 0 {
		return ""
	}

	// Scale to terminal width
	height := width * cropRows / cropCols / 2 // /2 because terminal chars are ~2x tall
	if height < 1 {
		height = 1
	}

	// Compute overlay grids if requested
	var borderGrid, lakeGrid [][]bool
	if showBorder || showLakes {
		borderGrid, lakeGrid = RenderOverlay(width, height, chMinLat, chMaxLat, chMinLon, chMaxLon)
	}

	toRune := precipToRune
	if noColor {
		toRune = precipToRuneNoColor
	}

	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcRow := startRow + y*cropRows/height
			srcCol := startCol + x*cropCols/width
			if srcRow >= grid.Rows || srcCol >= grid.Cols {
				sb.WriteString(" ")
				continue
			}
			val := grid.Data[srcRow*grid.Cols+srcCol]

			// Precipitation takes priority
			if val > 0 {
				sb.WriteString(toRune(val))
				continue
			}

			// Lake overlay
			if showLakes && lakeGrid != nil && lakeGrid[y][x] {
				if noColor {
					sb.WriteString("~")
				} else {
					sb.WriteString("\033[38;5;33m~\033[0m")
				}
				continue
			}

			// Border overlay
			if showBorder && borderGrid != nil && borderGrid[y][x] {
				if noColor {
					sb.WriteString("·")
				} else {
					sb.WriteString("\033[38;5;255m·\033[0m")
				}
				continue
			}

			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}

	// Add legend
	sb.WriteString("\n")
	if noColor {
		sb.WriteString("Legend: ░ <0.1mm  ▒ <0.5mm  ▓ <2mm  █ >2mm")
		if showBorder {
			sb.WriteString("  · border")
		}
		if showLakes {
			sb.WriteString("  ~ lake")
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("Legend: \033[38;5;39m░\033[0m <0.1mm  \033[38;5;33m▒\033[0m <0.5mm  \033[38;5;28m▓\033[0m <1mm  \033[38;5;226m▓\033[0m <2mm  \033[38;5;208m█\033[0m <5mm  \033[38;5;196m█\033[0m >5mm")
		if showBorder {
			sb.WriteString("  \033[38;5;255m·\033[0m border")
		}
		if showLakes {
			sb.WriteString("  \033[38;5;33m~\033[0m lake")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
