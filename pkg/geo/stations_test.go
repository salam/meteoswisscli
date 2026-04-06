package geo

import "testing"

func TestLookupStation_Valid(t *testing.T) {
	s := LookupStation("SMA")
	if s == nil {
		t.Fatal("LookupStation(SMA) should find Zurich Fluntern")
	}
	if s.Code != "SMA" {
		t.Errorf("Code = %q, want SMA", s.Code)
	}
	if s.Canton != "ZH" {
		t.Errorf("Canton = %q, want ZH", s.Canton)
	}
}

func TestLookupStation_CaseInsensitive(t *testing.T) {
	s := LookupStation("sma")
	if s == nil {
		t.Fatal("LookupStation(sma) should find station case-insensitively")
	}
	if s.Code != "SMA" {
		t.Errorf("Code = %q, want SMA", s.Code)
	}
}

func TestLookupStation_Invalid(t *testing.T) {
	s := LookupStation("XXXXXX")
	if s != nil {
		t.Errorf("LookupStation(XXXXXX) should return nil, got %+v", s)
	}
}

func TestLookupStation_EmptyCode(t *testing.T) {
	s := LookupStation("")
	if s != nil {
		t.Errorf("LookupStation('') should return nil, got %+v", s)
	}
}

func TestFindNearestStations_Zurich(t *testing.T) {
	// Zurich main station coordinates
	results, err := FindNearestStations(47.3769, 8.5417, 3)
	if err != nil {
		t.Fatalf("FindNearestStations error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	// First result should be the closest station
	if results[0].Distance > results[1].Distance {
		t.Errorf("results not sorted by distance: %.1f > %.1f", results[0].Distance, results[1].Distance)
	}
	if results[1].Distance > results[2].Distance {
		t.Errorf("results not sorted by distance: %.1f > %.1f", results[1].Distance, results[2].Distance)
	}
}

func TestFindNearestStations_LimitZero(t *testing.T) {
	results, err := FindNearestStations(47.3769, 8.5417, 0)
	if err != nil {
		t.Fatalf("FindNearestStations error = %v", err)
	}
	// With limit 0, should return all stations
	if len(results) < 10 {
		t.Errorf("expected many stations with limit 0, got %d", len(results))
	}
}

func TestResolveStation_ByCode(t *testing.T) {
	results, err := ResolveStation("SMA", 5)
	if err != nil {
		t.Fatalf("ResolveStation(SMA) error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for station code, got %d", len(results))
	}
	if results[0].Station.Code != "SMA" {
		t.Errorf("Code = %q, want SMA", results[0].Station.Code)
	}
	if results[0].Distance != 0 {
		t.Errorf("Distance = %f, want 0 for exact code match", results[0].Distance)
	}
}

func TestResolveStation_ByCoordinates(t *testing.T) {
	results, err := ResolveStation("47.37,8.55", 3)
	if err != nil {
		t.Fatalf("ResolveStation(coordinates) error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestResolveStation_ByPlaceName(t *testing.T) {
	results, err := ResolveStation("Bern", 3)
	if err != nil {
		t.Fatalf("ResolveStation(Bern) error = %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for Bern")
	}
}

func TestResolveStation_Invalid(t *testing.T) {
	_, err := ResolveStation("Nonexistentville", 3)
	if err == nil {
		t.Error("ResolveStation should fail for unknown place")
	}
}
