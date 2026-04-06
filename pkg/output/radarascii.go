package output

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
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

// ExtractRadarGrid uses Python+h5py to extract the precipitation grid from an HDF5 file.
// Returns the grid as a flat float64 array.
func ExtractRadarGrid(h5path string) (*RadarGrid, error) {
	script := `
import h5py, struct, sys, math
with h5py.File(sys.argv[1], 'r') as f:
    data = f['dataset1/data1/data'][()]
    where = f['where']
    rows, cols = data.shape
    ll_lat = float(where.attrs['LL_lat'])
    ll_lon = float(where.attrs['LL_lon'])
    ur_lat = float(where.attrs['UR_lat'])
    ur_lon = float(where.attrs['UR_lon'])
    # Write header: rows(4), cols(4), ll_lat(8), ll_lon(8), ur_lat(8), ur_lon(8)
    sys.stdout.buffer.write(struct.pack('<ii', rows, cols))
    sys.stdout.buffer.write(struct.pack('<dddd', ll_lat, ll_lon, ur_lat, ur_lon))
    # Write data row by row as float64
    for row in range(rows):
        for col in range(cols):
            v = float(data[row, col])
            if math.isnan(v) or math.isinf(v):
                v = 0.0
            sys.stdout.buffer.write(struct.pack('<d', v))
`
	cmd := exec.Command("python3", "-c", script, h5path)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("extract radar data (requires python3 + h5py): %w", err)
	}

	if len(out) < 40 {
		return nil, fmt.Errorf("radar data too small: %d bytes", len(out))
	}

	// Parse header
	rows := int(binary.LittleEndian.Uint32(out[0:4]))
	cols := int(binary.LittleEndian.Uint32(out[4:8]))
	llLat := math.Float64frombits(binary.LittleEndian.Uint64(out[8:16]))
	llLon := math.Float64frombits(binary.LittleEndian.Uint64(out[16:24]))
	urLat := math.Float64frombits(binary.LittleEndian.Uint64(out[24:32]))
	urLon := math.Float64frombits(binary.LittleEndian.Uint64(out[32:40]))

	expectedSize := 40 + rows*cols*8
	if len(out) < expectedSize {
		return nil, fmt.Errorf("radar data truncated: got %d, expected %d bytes", len(out), expectedSize)
	}

	data := make([]float64, rows*cols)
	for i := range data {
		offset := 40 + i*8
		data[i] = math.Float64frombits(binary.LittleEndian.Uint64(out[offset : offset+8]))
	}

	return &RadarGrid{
		Rows: rows, Cols: cols, Data: data,
		MinLat: llLat, MaxLat: urLat, MinLon: llLon, MaxLon: urLon,
	}, nil
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
func RenderRadarASCII(grid *RadarGrid, width int, noColor bool) string {
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
			sb.WriteString(toRune(val))
		}
		sb.WriteString("\n")
	}

	// Add legend
	sb.WriteString("\n")
	if noColor {
		sb.WriteString("Legend: ░ <0.1mm  ▒ <0.5mm  ▓ <2mm  █ >2mm\n")
	} else {
		sb.WriteString("Legend: \033[38;5;39m░\033[0m <0.1mm  \033[38;5;33m▒\033[0m <0.5mm  \033[38;5;28m▓\033[0m <1mm  \033[38;5;226m▓\033[0m <2mm  \033[38;5;208m█\033[0m <5mm  \033[38;5;196m█\033[0m >5mm\n")
	}

	return sb.String()
}
