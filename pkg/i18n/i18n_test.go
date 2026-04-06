package i18n

import "testing"

func TestT_German(t *testing.T) {
	old := Lang
	defer func() { Lang = old }()

	Lang = "de"
	tests := []struct {
		key  string
		want string
	}{
		{"Current Weather", "Aktuelles Wetter"},
		{"Forecast", "Vorhersage"},
		{"Warnings", "Warnungen"},
		{"STATION", "STATION"},
		{"WIND", "WIND"},
	}
	for _, tt := range tests {
		got := T(tt.key)
		if got != tt.want {
			t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestT_French(t *testing.T) {
	old := Lang
	defer func() { Lang = old }()

	Lang = "fr"
	tests := []struct {
		key  string
		want string
	}{
		{"Current Weather", "Météo actuelle"},
		{"Forecast", "Prévisions"},
		{"Warnings", "Alertes"},
		{"WIND", "VENT"},
	}
	for _, tt := range tests {
		got := T(tt.key)
		if got != tt.want {
			t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestT_Italian(t *testing.T) {
	old := Lang
	defer func() { Lang = old }()

	Lang = "it"
	tests := []struct {
		key  string
		want string
	}{
		{"Current Weather", "Meteo attuale"},
		{"STATION", "STAZIONE"},
		{"RAIN", "PIOGGIA"},
	}
	for _, tt := range tests {
		got := T(tt.key)
		if got != tt.want {
			t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestT_English(t *testing.T) {
	old := Lang
	defer func() { Lang = old }()

	Lang = "en"
	tests := []struct {
		key  string
		want string
	}{
		{"Current Weather", "Current Weather"},
		{"Forecast", "Forecast"},
		{"RAIN", "RAIN"},
	}
	for _, tt := range tests {
		got := T(tt.key)
		if got != tt.want {
			t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestT_FallbackForUnknownKey(t *testing.T) {
	old := Lang
	defer func() { Lang = old }()

	Lang = "de"
	got := T("this key does not exist")
	if got != "this key does not exist" {
		t.Errorf("T(unknown key) = %q, want key returned as-is", got)
	}
}

func TestT_FallbackForUnknownLanguage(t *testing.T) {
	old := Lang
	defer func() { Lang = old }()

	Lang = "xx"
	got := T("Current Weather")
	// Should return the key itself since "xx" is not in translations
	if got != "Current Weather" {
		t.Errorf("T with unknown lang = %q, want key as fallback", got)
	}
}
