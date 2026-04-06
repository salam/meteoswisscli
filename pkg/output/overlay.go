package output

import (
	_ "embed"
	"encoding/json"
)

//go:embed data/swiss_outline.json
var swissOutlineData []byte

type swissOutline struct {
	Border [][]float64            `json:"border"` // [[lon, lat], ...]
	Lakes  map[string][][]float64 `json:"lakes"`  // name -> [[lon, lat], ...]
}

var outline *swissOutline

func loadOutline() *swissOutline {
	if outline != nil {
		return outline
	}
	outline = &swissOutline{}
	json.Unmarshal(swissOutlineData, outline)
	return outline
}

// RenderOverlay creates boolean grids marking border and lake pixels.
// Returns two grids: borderGrid and lakeGrid (true = pixel is border/lake).
func RenderOverlay(width, height int, minLat, maxLat, minLon, maxLon float64) (borderGrid, lakeGrid [][]bool) {
	o := loadOutline()

	borderGrid = make([][]bool, height)
	lakeGrid = make([][]bool, height)
	for y := range borderGrid {
		borderGrid[y] = make([]bool, width)
		lakeGrid[y] = make([]bool, width)
	}

	// Draw border as connected line segments
	for i := 0; i < len(o.Border)-1; i++ {
		lon1, lat1 := o.Border[i][0], o.Border[i][1]
		lon2, lat2 := o.Border[i+1][0], o.Border[i+1][1]
		drawLine(borderGrid, width, height, minLat, maxLat, minLon, maxLon, lon1, lat1, lon2, lat2)
	}

	// Draw lakes as outlines
	for _, lakePoints := range o.Lakes {
		for i := 0; i < len(lakePoints)-1; i++ {
			lon1, lat1 := lakePoints[i][0], lakePoints[i][1]
			lon2, lat2 := lakePoints[i+1][0], lakePoints[i+1][1]
			drawLine(lakeGrid, width, height, minLat, maxLat, minLon, maxLon, lon1, lat1, lon2, lat2)
		}
	}

	return borderGrid, lakeGrid
}

func drawLine(grid [][]bool, width, height int, minLat, maxLat, minLon, maxLon, lon1, lat1, lon2, lat2 float64) {
	// Convert geo coords to pixel coords
	x1 := int((lon1 - minLon) / (maxLon - minLon) * float64(width))
	y1 := int((maxLat - lat1) / (maxLat - minLat) * float64(height))
	x2 := int((lon2 - minLon) / (maxLon - minLon) * float64(width))
	y2 := int((maxLat - lat2) / (maxLat - minLat) * float64(height))

	// Bresenham's line algorithm
	dx := iabs(x2 - x1)
	dy := iabs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		if x1 >= 0 && x1 < width && y1 >= 0 && y1 < height {
			grid[y1][x1] = true
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func iabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
