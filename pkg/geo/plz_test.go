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

func TestParsePLZ_Coordinates_NotYetSupported(t *testing.T) {
	_, err := ParsePLZ("47.37,8.55")
	if err == nil {
		t.Error("ParsePLZ with coordinates should return error until lookup is implemented")
	}
}
