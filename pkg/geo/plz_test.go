package geo

import "testing"

func TestParsePLZ_Numeric(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"8001", "800100"},
		{"3012", "301200"},
		{"1000", "100000"},
		{"800100", "800100"},
		{"12345", "123450"},
	}
	for _, tt := range tests {
		got, err := ParsePLZ(tt.input)
		if err != nil {
			t.Errorf("ParsePLZ(%q) error = %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParsePLZ(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParsePLZ_Invalid(t *testing.T) {
	_, err := ParsePLZ("abc")
	if err == nil {
		t.Error("ParsePLZ(abc) should return error for non-numeric, non-coordinate input")
	}
}

func TestParseCoordinates(t *testing.T) {
	lat, lon, ok := parseCoordinates("47.37,8.55")
	if !ok {
		t.Fatal("parseCoordinates should succeed")
	}
	if lat < 47.36 || lat > 47.38 {
		t.Errorf("lat = %f, want ~47.37", lat)
	}
	if lon < 8.54 || lon > 8.56 {
		t.Errorf("lon = %f, want ~8.55", lon)
	}
}

func TestParseCoordinates_Invalid(t *testing.T) {
	_, _, ok := parseCoordinates("notcoords")
	if ok {
		t.Error("parseCoordinates should fail for non-coordinate input")
	}
}

func TestIsNumeric(t *testing.T) {
	if !isNumeric("8001") {
		t.Error("8001 should be numeric")
	}
	if isNumeric("Zürich") {
		t.Error("Zürich should not be numeric")
	}
}

func TestParsePLZ_Coordinates(t *testing.T) {
	// 47.37,8.55 is near Zürich (PLZ 8001)
	plz, err := ParsePLZ("47.37,8.55")
	if err != nil {
		t.Fatalf("ParsePLZ(coordinates) error = %v", err)
	}
	if plz == "" {
		t.Error("ParsePLZ(coordinates) should return a PLZ")
	}
}

func TestSearchLocation(t *testing.T) {
	loc, err := SearchLocation("Arosa GR")
	if err != nil {
		t.Fatalf("SearchLocation(Arosa GR) error = %v", err)
	}
	if loc.PLZ != "7050" {
		t.Errorf("PLZ = %q, want 7050", loc.PLZ)
	}
	if loc.Kanton != "GR" {
		t.Errorf("Kanton = %q, want GR", loc.Kanton)
	}
}

func TestSearchLocation_CityOnly(t *testing.T) {
	loc, err := SearchLocation("Zürich")
	if err != nil {
		t.Fatalf("SearchLocation(Zürich) error = %v", err)
	}
	if loc.Kanton != "ZH" {
		t.Errorf("Kanton = %q, want ZH", loc.Kanton)
	}
}

func TestFindNearest(t *testing.T) {
	// Coordinates near Arosa
	loc, err := FindNearest(46.78, 9.68)
	if err != nil {
		t.Fatalf("FindNearest error = %v", err)
	}
	if loc.PLZ != "7050" {
		t.Errorf("PLZ = %q, want 7050 (Arosa)", loc.PLZ)
	}
}
