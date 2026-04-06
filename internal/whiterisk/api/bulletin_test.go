package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBulletin(t *testing.T) {
	response := BulletinResponse{
		Bulletins: []Bulletin{
			{
				BulletinID:      "test-123",
				PublicationTime: "2026-04-05T17:00:00Z",
				ValidTime: ValidTime{
					StartTime: "2026-04-05T17:00:00Z",
					EndTime:   "2026-04-06T17:00:00Z",
				},
				Regions: []Region{
					{RegionID: "CH-7231", Name: "Davos"},
				},
				DangerRatings: []DangerRating{
					{MainValue: "considerable", Elevation: ElevationRange{UpperBound: 2200}},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bulletin/caaml/en/json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClientWithBase(server.URL, "en")
	result, err := c.GetBulletin()
	if err != nil {
		t.Fatalf("GetBulletin error = %v", err)
	}
	if len(result.Bulletins) != 1 {
		t.Fatalf("bulletins = %d, want 1", len(result.Bulletins))
	}
	if result.Bulletins[0].Regions[0].Name != "Davos" {
		t.Errorf("region = %q, want Davos", result.Bulletins[0].Regions[0].Name)
	}
}

func TestDangerLevelDisplay(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"low", "1 — Low"},
		{"moderate", "2 — Moderate"},
		{"considerable", "3 — Considerable"},
		{"high", "4 — High"},
		{"very_high", "5 — Very High"},
		{"unknown", "unknown"},
		{"", ""},
	}
	for _, tt := range tests {
		got := DangerLevelDisplay(tt.input)
		if got != tt.want {
			t.Errorf("DangerLevelDisplay(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestElevationRange_LowerBoundStr(t *testing.T) {
	tests := []struct {
		name  string
		elev  ElevationRange
		want  string
	}{
		{"nil lower", ElevationRange{LowerBound: nil}, ""},
		{"int lower", ElevationRange{LowerBound: 2200}, "2200"},
		{"float lower", ElevationRange{LowerBound: 2200.0}, "2200"},
		{"string lower", ElevationRange{LowerBound: "2200"}, "2200"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.elev.LowerBoundStr()
			if got != tt.want {
				t.Errorf("LowerBoundStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestElevationRange_UpperBoundStr(t *testing.T) {
	tests := []struct {
		name  string
		elev  ElevationRange
		want  string
	}{
		{"nil upper", ElevationRange{UpperBound: nil}, ""},
		{"int upper", ElevationRange{UpperBound: 2200}, "2200"},
		{"float upper", ElevationRange{UpperBound: 2200.0}, "2200"},
		{"string upper", ElevationRange{UpperBound: "3000"}, "3000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.elev.UpperBoundStr()
			if got != tt.want {
				t.Errorf("UpperBoundStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidTime(t *testing.T) {
	vt := ValidTime{
		StartTime: "2026-04-05T17:00:00Z",
		EndTime:   "2026-04-06T17:00:00Z",
	}
	if vt.StartTime == "" {
		t.Error("StartTime should not be empty")
	}
	if vt.EndTime == "" {
		t.Error("EndTime should not be empty")
	}
}
