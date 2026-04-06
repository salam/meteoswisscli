package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir, "testapp")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Lang != "" {
		t.Errorf("Lang = %q, want empty", cfg.Lang)
	}
	if len(cfg.Favorites) != 0 {
		t.Errorf("Favorites = %d, want 0", len(cfg.Favorites))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		Lang: "fr",
		Favorites: []Favorite{
			{Name: "Zürich", PLZ: "800100"},
		},
	}
	cfg.path = filepath.Join(dir, "testapp", "config.json")

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(dir, "testapp")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Lang != "fr" {
		t.Errorf("Lang = %q, want %q", loaded.Lang, "fr")
	}
	if len(loaded.Favorites) != 1 || loaded.Favorites[0].Name != "Zürich" {
		t.Errorf("Favorites not loaded correctly: %+v", loaded.Favorites)
	}
}

func TestAddRemoveFavorite(t *testing.T) {
	cfg := &Config{}
	cfg.AddFavorite(Favorite{Name: "Bern", PLZ: "301200"})
	if len(cfg.Favorites) != 1 {
		t.Fatalf("AddFavorite: len = %d, want 1", len(cfg.Favorites))
	}

	cfg.AddFavorite(Favorite{Name: "Basel", PLZ: "400000"})
	if len(cfg.Favorites) != 2 {
		t.Fatalf("AddFavorite: len = %d, want 2", len(cfg.Favorites))
	}

	cfg.RemoveFavorite("Bern")
	if len(cfg.Favorites) != 1 {
		t.Fatalf("RemoveFavorite: len = %d, want 1", len(cfg.Favorites))
	}
	if cfg.Favorites[0].Name != "Basel" {
		t.Errorf("remaining favorite = %q, want Basel", cfg.Favorites[0].Name)
	}
}

func TestDetectLang(t *testing.T) {
	os.Setenv("LANG", "fr_CH.UTF-8")
	defer os.Unsetenv("LANG")
	if got := DetectLang(""); got != "fr" {
		t.Errorf("DetectLang() = %q, want %q", got, "fr")
	}
}

func TestDetectLangOverride(t *testing.T) {
	if got := DetectLang("it"); got != "it" {
		t.Errorf("DetectLang(it) = %q, want %q", got, "it")
	}
}

func TestDetectLangEnvVar(t *testing.T) {
	os.Setenv("METEOSWISS_LANG", "en")
	defer os.Unsetenv("METEOSWISS_LANG")
	if got := DetectLangWithEnv("", "METEOSWISS_LANG"); got != "en" {
		t.Errorf("DetectLangWithEnv() = %q, want %q", got, "en")
	}
}
