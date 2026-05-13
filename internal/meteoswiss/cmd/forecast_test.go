package cmd

import (
	"strings"
	"testing"

	"github.com/salam/swissmeteocli/internal/meteoswiss/api"
	"github.com/salam/swissmeteocli/pkg/i18n"
)

func TestBuildComparisonTable_AlignsByDate(t *testing.T) {
	oldLang := i18n.Lang
	i18n.Lang = "en"
	defer func() { i18n.Lang = oldLang }()

	loc1 := []api.ForecastDay{
		{DayDate: "2026-05-13", IconDay: 2, TemperatureMin: 6, TemperatureMax: 16, Precipitation: 0.0},
		{DayDate: "2026-05-14", IconDay: 13, TemperatureMin: 6, TemperatureMax: 13, Precipitation: 2.5},
	}
	loc2 := []api.ForecastDay{
		{DayDate: "2026-05-13", IconDay: 7, TemperatureMin: 8, TemperatureMax: 14, Precipitation: 2.1},
		{DayDate: "2026-05-14", IconDay: 3, TemperatureMin: 9, TemperatureMax: 15, Precipitation: 0.5},
	}

	headers, rows := buildComparisonTable([][]api.ForecastDay{loc1, loc2})

	if len(headers) != 1+2*4 {
		t.Fatalf("headers len = %d, want %d", len(headers), 1+2*4)
	}
	if headers[0] != "DATE" {
		t.Errorf("headers[0] = %q, want DATE", headers[0])
	}
	if !strings.HasSuffix(headers[1], "1") || !strings.HasSuffix(headers[5], "2") {
		t.Errorf("expected '… 1' and '… 2' suffixes, got %v", headers)
	}

	if len(rows) != 2 {
		t.Fatalf("rows len = %d, want 2", len(rows))
	}
	if rows[0][0] != "2026-05-13" || rows[1][0] != "2026-05-14" {
		t.Errorf("rows dates = %v %v, want 2026-05-13 2026-05-14", rows[0][0], rows[1][0])
	}
	if rows[0][1] != "Mostly sunny" {
		t.Errorf("rows[0] loc1 icon = %q, want 'Mostly sunny'", rows[0][1])
	}
	if rows[0][5] != "Rain" {
		t.Errorf("rows[0] loc2 icon = %q, want 'Rain'", rows[0][5])
	}
}

func TestBuildComparisonTable_DisjointDates(t *testing.T) {
	loc1 := []api.ForecastDay{
		{DayDate: "2026-05-13", IconDay: 2, TemperatureMin: 6, TemperatureMax: 16},
	}
	loc2 := []api.ForecastDay{
		{DayDate: "2026-05-14", IconDay: 3, TemperatureMin: 9, TemperatureMax: 15},
	}

	_, rows := buildComparisonTable([][]api.ForecastDay{loc1, loc2})

	if len(rows) != 2 {
		t.Fatalf("rows len = %d, want 2 (union of dates)", len(rows))
	}
	if rows[0][0] != "2026-05-13" || rows[1][0] != "2026-05-14" {
		t.Errorf("rows should be sorted by date, got %v %v", rows[0][0], rows[1][0])
	}
	if rows[0][1] != "Mostly sunny" {
		t.Errorf("rows[0] loc1 icon = %q, want present", rows[0][1])
	}
	if rows[0][5] != "—" {
		t.Errorf("rows[0] loc2 icon = %q, want '—' (blank cell)", rows[0][5])
	}
	if rows[1][1] != "—" {
		t.Errorf("rows[1] loc1 icon = %q, want '—' (blank cell)", rows[1][1])
	}
	if rows[1][5] != "Partly cloudy" {
		t.Errorf("rows[1] loc2 icon = %q, want 'Partly cloudy'", rows[1][5])
	}
}
