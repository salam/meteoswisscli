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
					{MainValue: "considerable", Elevation: ElevationRange{UpperBound: "treeline"}},
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
	if got := DangerLevelDisplay("considerable"); got != "3 — Considerable" {
		t.Errorf("got %q, want %q", got, "3 — Considerable")
	}
	if got := DangerLevelDisplay("unknown"); got != "unknown" {
		t.Errorf("got %q, want %q", got, "unknown")
	}
}
