package output

import (
	"math"
	"strings"
	"testing"
)

func TestPrecipToRuneNoColor_Zero(t *testing.T) {
	got := precipToRuneNoColor(0)
	if got != " " {
		t.Errorf("precipToRuneNoColor(0) = %q, want space", got)
	}
}

func TestPrecipToRuneNoColor_Negative(t *testing.T) {
	got := precipToRuneNoColor(-1)
	if got != " " {
		t.Errorf("precipToRuneNoColor(-1) = %q, want space", got)
	}
}

func TestPrecipToRuneNoColor_Trace(t *testing.T) {
	got := precipToRuneNoColor(0.05)
	if got != "░" {
		t.Errorf("precipToRuneNoColor(0.05) = %q, want ░", got)
	}
}

func TestPrecipToRuneNoColor_Light(t *testing.T) {
	got := precipToRuneNoColor(0.3)
	if got != "▒" {
		t.Errorf("precipToRuneNoColor(0.3) = %q, want ▒", got)
	}
}

func TestPrecipToRuneNoColor_Moderate(t *testing.T) {
	got := precipToRuneNoColor(1.0)
	if got != "▓" {
		t.Errorf("precipToRuneNoColor(1.0) = %q, want ▓", got)
	}
}

func TestPrecipToRuneNoColor_Heavy(t *testing.T) {
	got := precipToRuneNoColor(5.0)
	if got != "█" {
		t.Errorf("precipToRuneNoColor(5.0) = %q, want █", got)
	}
}

func TestPrecipToRune_ZeroReturnsSpace(t *testing.T) {
	got := precipToRune(0)
	if got != " " {
		t.Errorf("precipToRune(0) = %q, want space", got)
	}
}

func TestPrecipToRune_NonZeroHasEscape(t *testing.T) {
	got := precipToRune(1.0)
	if !strings.Contains(got, "\033[") {
		t.Errorf("precipToRune(1.0) should contain ANSI escape, got %q", got)
	}
}

func TestRenderRadarASCII_EmptyGrid(t *testing.T) {
	grid := &RadarGrid{Rows: 0, Cols: 0}
	result := RenderRadarASCII(grid, 80, true, false, false)
	if result != "" {
		t.Errorf("empty grid should produce empty output, got %q", result)
	}
}

func TestRenderRadarASCII_SyntheticGrid(t *testing.T) {
	// Create a small grid covering Switzerland
	rows, cols := 100, 100
	data := make([]float64, rows*cols)
	// Fill with some precipitation in the middle
	for r := 40; r < 60; r++ {
		for c := 40; c < 60; c++ {
			data[r*cols+c] = 1.5
		}
	}

	grid := &RadarGrid{
		Rows:   rows,
		Cols:   cols,
		Data:   data,
		MinLat: 44.0,
		MaxLat: 50.0,
		MinLon: 4.0,
		MaxLon: 12.0,
	}

	result := RenderRadarASCII(grid, 40, true, false, false)
	if result == "" {
		t.Error("RenderRadarASCII should produce output for valid grid")
	}
	if !strings.Contains(result, "Legend:") {
		t.Error("output should contain legend")
	}
}

func TestRenderRadarASCII_NoColorFlag(t *testing.T) {
	rows, cols := 50, 50
	data := make([]float64, rows*cols)
	for i := range data {
		data[i] = 0.5
	}

	grid := &RadarGrid{
		Rows: rows, Cols: cols, Data: data,
		MinLat: 44.0, MaxLat: 50.0, MinLon: 4.0, MaxLon: 12.0,
	}

	result := RenderRadarASCII(grid, 30, true, false, false)
	if strings.Contains(result, "\033[") {
		t.Error("noColor=true should not contain ANSI escape codes")
	}
}

func TestRenderRadarASCII_WithColorFlag(t *testing.T) {
	rows, cols := 50, 50
	data := make([]float64, rows*cols)
	for i := range data {
		data[i] = 0.5
	}

	grid := &RadarGrid{
		Rows: rows, Cols: cols, Data: data,
		MinLat: 44.0, MaxLat: 50.0, MinLon: 4.0, MaxLon: 12.0,
	}

	result := RenderRadarASCII(grid, 30, false, false, false)
	if !strings.Contains(result, "\033[") {
		t.Error("noColor=false should contain ANSI escape codes for non-zero precip")
	}
}

func TestRenderRadarASCII_NaNData(t *testing.T) {
	rows, cols := 50, 50
	data := make([]float64, rows*cols)
	for i := range data {
		data[i] = math.NaN()
	}

	grid := &RadarGrid{
		Rows: rows, Cols: cols, Data: data,
		MinLat: 44.0, MaxLat: 50.0, MinLon: 4.0, MaxLon: 12.0,
	}

	// Should not panic
	result := RenderRadarASCII(grid, 30, true, false, false)
	if result == "" {
		t.Error("should produce output even with NaN data")
	}
}

func TestRenderRadarASCII_WithBorder(t *testing.T) {
	rows, cols := 50, 50
	data := make([]float64, rows*cols)

	grid := &RadarGrid{
		Rows: rows, Cols: cols, Data: data,
		MinLat: 44.0, MaxLat: 50.0, MinLon: 4.0, MaxLon: 12.0,
	}

	result := RenderRadarASCII(grid, 30, true, true, false)
	if !strings.Contains(result, "border") {
		t.Error("legend should mention border when showBorder=true")
	}
}
