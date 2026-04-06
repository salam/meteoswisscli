package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Lang            string     `json:"lang,omitempty"`
	DefaultLocation string     `json:"default_location,omitempty"`
	Favorites       []Favorite `json:"favorites,omitempty"`
	path            string
}

type Favorite struct {
	Name    string  `json:"name"`
	PLZ     string  `json:"plz,omitempty"`
	Region  string  `json:"region,omitempty"`
	Station string  `json:"station,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lon     float64 `json:"lon,omitempty"`
}

func Load(baseDir, appName string) (*Config, error) {
	p := filepath.Join(baseDir, appName, "config.json")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{path: p}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.path = p
	return &cfg, nil
}

func LoadDefault(appName string) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return Load(filepath.Join(home, ".config"), appName)
}

func (c *Config) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0600)
}

func (c *Config) Clear() error {
	return os.Remove(c.path)
}

func (c *Config) AddFavorite(f Favorite) {
	c.Favorites = append(c.Favorites, f)
}

func (c *Config) RemoveFavorite(name string) {
	filtered := c.Favorites[:0]
	for _, f := range c.Favorites {
		if !strings.EqualFold(f.Name, name) {
			filtered = append(filtered, f)
		}
	}
	c.Favorites = filtered
}

func DetectLang(flagOverride string) string {
	return DetectLangWithEnv(flagOverride, "")
}

func DetectLangWithEnv(flagOverride, envKey string) string {
	if flagOverride != "" {
		return flagOverride
	}
	if envKey != "" {
		if v := os.Getenv(envKey); v != "" {
			return v
		}
	}
	for _, key := range []string{"LANG", "LC_ALL"} {
		val := os.Getenv(key)
		if val == "" || val == "C" || val == "POSIX" {
			continue
		}
		lang := strings.SplitN(val, "_", 2)[0]
		switch lang {
		case "de", "fr", "it", "en":
			return lang
		}
	}
	return "de"
}
