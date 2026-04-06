package geo

import "testing"

func TestSearchLocation_ExactName(t *testing.T) {
	loc, err := SearchLocation("Bern")
	if err != nil {
		t.Fatalf("SearchLocation(Bern) error = %v", err)
	}
	if loc.Kanton != "BE" {
		t.Errorf("Kanton = %q, want BE", loc.Kanton)
	}
	if loc.Name != "Bern" {
		t.Errorf("Name = %q, want Bern", loc.Name)
	}
}

func TestSearchLocation_NameWithCanton(t *testing.T) {
	loc, err := SearchLocation("Basel BS")
	if err != nil {
		t.Fatalf("SearchLocation(Basel BS) error = %v", err)
	}
	if loc.Kanton != "BS" {
		t.Errorf("Kanton = %q, want BS", loc.Kanton)
	}
}

func TestSearchLocation_PartialMatch(t *testing.T) {
	loc, err := SearchLocation("Luga")
	if err != nil {
		t.Fatalf("SearchLocation(Luga) error = %v", err)
	}
	if loc.Kanton != "TI" {
		t.Errorf("Kanton = %q, want TI", loc.Kanton)
	}
}

func TestSearchLocation_CaseInsensitive(t *testing.T) {
	loc, err := SearchLocation("zürich")
	if err != nil {
		t.Fatalf("SearchLocation(zürich) error = %v", err)
	}
	if loc.Kanton != "ZH" {
		t.Errorf("Kanton = %q, want ZH", loc.Kanton)
	}
}

func TestSearchLocation_NotFound(t *testing.T) {
	_, err := SearchLocation("Nonexistentville")
	if err == nil {
		t.Error("SearchLocation should return error for unknown location")
	}
}

func TestSearchLocation_Whitespace(t *testing.T) {
	loc, err := SearchLocation("  Bern  ")
	if err != nil {
		t.Fatalf("SearchLocation with whitespace error = %v", err)
	}
	if loc.Kanton != "BE" {
		t.Errorf("Kanton = %q, want BE", loc.Kanton)
	}
}

func TestFindNearest_Zurich(t *testing.T) {
	// Coordinates of Zurich main station
	loc, err := FindNearest(47.3769, 8.5417)
	if err != nil {
		t.Fatalf("FindNearest error = %v", err)
	}
	if loc.Kanton != "ZH" {
		t.Errorf("Kanton = %q, want ZH for Zurich coordinates", loc.Kanton)
	}
}

func TestFindNearest_Geneva(t *testing.T) {
	// Coordinates near Geneva
	loc, err := FindNearest(46.2044, 6.1432)
	if err != nil {
		t.Fatalf("FindNearest error = %v", err)
	}
	if loc.Kanton != "GE" {
		t.Errorf("Kanton = %q, want GE for Geneva coordinates", loc.Kanton)
	}
}

func TestFindNearest_Lugano(t *testing.T) {
	// Coordinates near Lugano
	loc, err := FindNearest(46.0037, 8.9511)
	if err != nil {
		t.Fatalf("FindNearest error = %v", err)
	}
	if loc.Kanton != "TI" {
		t.Errorf("Kanton = %q, want TI for Lugano coordinates", loc.Kanton)
	}
}
