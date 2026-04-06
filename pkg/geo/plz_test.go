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

func TestPadPLZ(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"8001", "800100"},
		{"80", "800000"},
		{"800100", "800100"},
		{"1", "100000"},
		{"123456", "123456"},
		{"1234567", "123456"}, // truncates to 6
		{"", "000000"},
	}
	for _, tt := range tests {
		got := padPLZ(tt.input)
		if got != tt.want {
			t.Errorf("padPLZ(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseCoordinates_WithSpaces(t *testing.T) {
	lat, lon, ok := parseCoordinates(" 47.37 , 8.55 ")
	if !ok {
		t.Fatal("parseCoordinates should succeed with spaces")
	}
	if lat < 47.36 || lat > 47.38 {
		t.Errorf("lat = %f, want ~47.37", lat)
	}
	if lon < 8.54 || lon > 8.56 {
		t.Errorf("lon = %f, want ~8.55", lon)
	}
}

func TestParseCoordinates_NegativeCoords(t *testing.T) {
	lat, lon, ok := parseCoordinates("-33.86,151.20")
	if !ok {
		t.Fatal("parseCoordinates should succeed with negative coords")
	}
	if lat > -33.85 || lat < -33.87 {
		t.Errorf("lat = %f, want ~-33.86", lat)
	}
	if lon < 151.19 || lon > 151.21 {
		t.Errorf("lon = %f, want ~151.20", lon)
	}
}

func TestParseCoordinates_OnlyOneValue(t *testing.T) {
	_, _, ok := parseCoordinates("47.37")
	if ok {
		t.Error("parseCoordinates should fail with only one value")
	}
}

func TestParseCoordinates_EmptyString(t *testing.T) {
	_, _, ok := parseCoordinates("")
	if ok {
		t.Error("parseCoordinates should fail with empty string")
	}
}

func TestParseCoordinates_NonNumeric(t *testing.T) {
	_, _, ok := parseCoordinates("abc,def")
	if ok {
		t.Error("parseCoordinates should fail with non-numeric input")
	}
}

func TestIsNumeric_Empty(t *testing.T) {
	if isNumeric("") {
		t.Error("empty string should not be numeric")
	}
}

func TestIsNumeric_WithDot(t *testing.T) {
	if isNumeric("3.14") {
		t.Error("3.14 should not be numeric (contains dot)")
	}
}

func TestResolvePLZ_ByPLZ(t *testing.T) {
	r, err := ResolvePLZ("8001")
	if err != nil {
		t.Fatalf("ResolvePLZ(8001) error = %v", err)
	}
	if r.PLZ != "800100" {
		t.Errorf("PLZ = %q, want 800100", r.PLZ)
	}
	// Should have resolved location metadata
	if r.Location != nil && r.Location.Kanton != "ZH" {
		t.Errorf("Kanton = %q, want ZH", r.Location.Kanton)
	}
}

func TestResolvePLZ_ByName(t *testing.T) {
	r, err := ResolvePLZ("Bern")
	if err != nil {
		t.Fatalf("ResolvePLZ(Bern) error = %v", err)
	}
	if r.Location == nil {
		t.Fatal("Location should not be nil for name resolution")
	}
	if r.Location.Kanton != "BE" {
		t.Errorf("Kanton = %q, want BE", r.Location.Kanton)
	}
}

func TestResolvePLZ_ByCoordinates(t *testing.T) {
	r, err := ResolvePLZ("46.95,7.45")
	if err != nil {
		t.Fatalf("ResolvePLZ(coordinates) error = %v", err)
	}
	if r.Location == nil {
		t.Fatal("Location should not be nil for coordinate resolution")
	}
	// Should be near Bern
	if r.Location.Kanton != "BE" {
		t.Errorf("Kanton = %q, want BE", r.Location.Kanton)
	}
}

func TestResolvedLocation_Label(t *testing.T) {
	// With location
	r := ResolvedLocation{
		PLZ:      "800100",
		Location: &Location{PLZ: "8001", Name: "Zürich", Kanton: "ZH"},
	}
	label := r.Label()
	if label != "8001 Zürich ZH" {
		t.Errorf("Label() = %q, want %q", label, "8001 Zürich ZH")
	}

	// Without location
	r2 := ResolvedLocation{PLZ: "800100"}
	if r2.Label() != "800100" {
		t.Errorf("Label() = %q, want %q", r2.Label(), "800100")
	}
}

func TestHaversineDistance(t *testing.T) {
	// Zurich to Bern is roughly 95 km
	d := haversineDistance(47.3769, 8.5417, 46.9480, 7.4474)
	if d < 80 || d > 110 {
		t.Errorf("haversineDistance(Zurich, Bern) = %.1f km, expected ~95 km", d)
	}

	// Same point should be 0
	d0 := haversineDistance(47.0, 8.0, 47.0, 8.0)
	if d0 != 0 {
		t.Errorf("haversineDistance(same point) = %f, want 0", d0)
	}
}
