# SwissCLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build two Go CLI tools (`meteoswiss` and `whiterisk`) for Swiss weather and avalanche data, sharing output/config/geo packages.

**Architecture:** Monorepo with `cmd/meteoswiss/` and `cmd/whiterisk/` entry points, `internal/` for CLI-specific code, `pkg/` for shared packages. All APIs are unauthenticated. Follows the klapp CLI pattern (cobra, one file per command, TTY tables vs JSON pipes).

**Tech Stack:** Go, cobra, modernc.org/sqlite, stdlib (net/http, encoding/json, encoding/csv, text/tabwriter, image/png, image/jpeg, database/sql)

**Spec:** `docs/superpowers/specs/2026-04-05-swisscli-design.md`

---

## Phase 1: Project Scaffolding & Shared Packages

### Task 1: Initialize Go module and project structure

**Files:**
- Create: `go.mod`
- Create: `cmd/meteoswiss/main.go`
- Create: `cmd/whiterisk/main.go`
- Create: `Makefile`

- [ ] **Step 1: Initialize the Go module**

```bash
cd /Users/matthias/Development/meteoswisscli
git init
go mod init github.com/matthias/swisscli
```

- [ ] **Step 2: Create meteoswiss entry point**

Create `cmd/meteoswiss/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("meteoswiss")
}
```

- [ ] **Step 3: Create whiterisk entry point**

Create `cmd/whiterisk/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("whiterisk")
}
```

- [ ] **Step 4: Create Makefile**

Create `Makefile`:

```makefile
MODULE  := $(shell head -1 go.mod | awk '{print $$2}')
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w

.PHONY: build clean all

build:
	go build -ldflags "$(LDFLAGS) -X '$(MODULE)/internal/meteoswiss/cmd.version=$(VERSION)'" -o meteoswiss ./cmd/meteoswiss
	go build -ldflags "$(LDFLAGS) -X '$(MODULE)/internal/whiterisk/cmd.version=$(VERSION)'" -o whiterisk ./cmd/whiterisk

clean:
	rm -f meteoswiss whiterisk
	rm -rf dist

all: clean
	@for app in meteoswiss whiterisk; do \
		for os_arch in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64; do \
			GOOS=$${os_arch%/*} GOARCH=$${os_arch#*/} \
			go build -ldflags "$(LDFLAGS) -X '$(MODULE)/internal/$${app}/cmd.version=$(VERSION)'" \
				-o dist/$${app}-$${os_arch%/*}-$${os_arch#*/}$$([ "$${os_arch%/*}" = windows ] && echo .exe) \
				./cmd/$${app}; \
		done \
	done
```

- [ ] **Step 5: Verify both binaries build**

```bash
go build ./cmd/meteoswiss && ./meteoswiss
go build ./cmd/whiterisk && ./whiterisk
```

Expected: each prints its name.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: initialize Go module with two entry points and Makefile"
```

---

### Task 2: pkg/source — Attribution strings

**Files:**
- Create: `pkg/source/source.go`
- Create: `pkg/source/source_test.go`

- [ ] **Step 1: Write the test**

Create `pkg/source/source_test.go`:

```go
package source

import "testing"

func TestMeteoSwissAttribution(t *testing.T) {
	want := "Quelle: MeteoSchweiz; Source: MétéoSuisse; Fonte: MeteoSvizzera; Source: MeteoSwiss"
	if MeteoSwiss != want {
		t.Errorf("MeteoSwiss = %q, want %q", MeteoSwiss, want)
	}
}

func TestSLFAttribution(t *testing.T) {
	want := "Quelle: SLF/WSL; Source: SLF/WSL"
	if SLF != want {
		t.Errorf("SLF = %q, want %q", SLF, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/source/
```

Expected: FAIL — package does not exist.

- [ ] **Step 3: Write implementation**

Create `pkg/source/source.go`:

```go
package source

const MeteoSwiss = "Quelle: MeteoSchweiz; Source: MétéoSuisse; Fonte: MeteoSvizzera; Source: MeteoSwiss"
const SLF = "Quelle: SLF/WSL; Source: SLF/WSL"
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./pkg/source/
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/source/
git commit -m "feat: add attribution string constants for MeteoSwiss and SLF"
```

---

### Task 3: pkg/output — Core output functions

**Files:**
- Create: `pkg/output/output.go`
- Create: `pkg/output/output_test.go`

- [ ] **Step 1: Write tests**

Create `pkg/output/output_test.go`:

```go
package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestIsInteractive_ForceJSON(t *testing.T) {
	ForceJSON = true
	defer func() { ForceJSON = false }()
	if IsInteractive() {
		t.Error("IsInteractive() should return false when ForceJSON is true")
	}
}

func TestTable(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Table([]string{"NAME", "VALUE"}, [][]string{{"temp", "20°C"}})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "NAME") {
		t.Errorf("Table output should contain header NAME, got: %s", out)
	}
	if !strings.Contains(out, "temp") {
		t.Errorf("Table output should contain row data, got: %s", out)
	}
}

func TestJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	JSON(map[string]string{"key": "value"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON output should be valid JSON: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("JSON key = %q, want %q", result["key"], "value")
	}
}

func TestError(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Error("something broke")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if !strings.Contains(buf.String(), "something broke") {
		t.Errorf("Error should write to stderr, got: %s", buf.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./pkg/output/
```

Expected: FAIL.

- [ ] **Step 3: Write implementation**

Create `pkg/output/output.go`:

```go
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

var ForceJSON bool
var NoColor bool

func IsInteractive() bool {
	if ForceJSON {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if !NoColor {
		fmt.Fprintf(w, "\033[1m%s\033[0m\n", strings.Join(headers, "\t"))
	} else {
		fmt.Fprintln(w, strings.Join(headers, "\t"))
	}
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

func JSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func Section(title string) {
	fmt.Printf("\n--- %s ---\n", title)
}

func Error(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}

func JSONError(msg string) {
	JSON(map[string]string{"error": msg})
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/output/
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/output/
git commit -m "feat: add output package with table, JSON, error, and TTY detection"
```

---

### Task 4: pkg/output — Browser, image save, and ASCII art

**Files:**
- Create: `pkg/output/browser.go`
- Create: `pkg/output/ascii.go`
- Create: `pkg/output/ascii_test.go`

- [ ] **Step 1: Write ASCII art test**

Create `pkg/output/ascii_test.go`:

```go
package output

import (
	"image"
	"image/color"
	"testing"
)

func TestPixelToBlock(t *testing.T) {
	tests := []struct {
		brightness uint8
		want       rune
	}{
		{0, ' '},
		{64, '░'},
		{128, '▒'},
		{192, '▓'},
		{255, '█'},
	}
	for _, tt := range tests {
		got := pixelToBlock(tt.brightness)
		if got != tt.want {
			t.Errorf("pixelToBlock(%d) = %c, want %c", tt.brightness, got, tt.want)
		}
	}
}

func TestRenderASCII(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetGray(x, y, color.Gray{Y: 200})
		}
	}
	result := RenderASCII(img, 4)
	if len(result) == 0 {
		t.Error("RenderASCII should produce output")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/output/ -run "Pixel|Render"
```

Expected: FAIL.

- [ ] **Step 3: Write ASCII art implementation**

Create `pkg/output/ascii.go`:

```go
package output

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

func pixelToBlock(brightness uint8) rune {
	switch {
	case brightness < 51:
		return ' '
	case brightness < 102:
		return '░'
	case brightness < 153:
		return '▒'
	case brightness < 204:
		return '▓'
	default:
		return '█'
	}
}

func RenderASCII(img image.Image, width int) string {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	if imgW == 0 || imgH == 0 {
		return ""
	}

	// Each terminal character is roughly twice as tall as wide
	height := width * imgH / imgW / 2
	if height < 1 {
		height = 1
	}

	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := bounds.Min.X + x*imgW/width
			srcY := bounds.Min.Y + y*imgH/height
			r, g, b, _ := img.At(srcX, srcY).RGBA()
			brightness := uint8((r/256*299 + g/256*587 + b/256*114) / 1000)
			sb.WriteRune(pixelToBlock(brightness))
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

func ASCIIMap(imageURL string, width int) error {
	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	if width <= 0 {
		width = 80
	}
	fmt.Print(RenderASCII(img, width))
	return nil
}

func SaveImage(imageURL string, path string) error {
	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Write browser opener**

Create `pkg/output/browser.go`:

```go
package output

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %s — open manually: %s", runtime.GOOS, url)
	}
	return cmd.Start()
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./pkg/output/
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/output/
git commit -m "feat: add browser opener, image save, and ASCII art rendering"
```

---

### Task 5: pkg/config — Configuration management

**Files:**
- Create: `pkg/config/config.go`
- Create: `pkg/config/config_test.go`

- [ ] **Step 1: Write tests**

Create `pkg/config/config_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./pkg/config/
```

Expected: FAIL.

- [ ] **Step 3: Write implementation**

Create `pkg/config/config.go`:

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Lang      string     `json:"lang,omitempty"`
	Favorites []Favorite `json:"favorites,omitempty"`
	path      string
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
		// Extract language code: "de_CH.UTF-8" → "de"
		lang := strings.SplitN(val, "_", 2)[0]
		switch lang {
		case "de", "fr", "it", "en":
			return lang
		}
	}
	return "de"
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./pkg/config/
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/config/
git commit -m "feat: add config package with load/save, favorites, and language detection"
```

---

### Task 6: pkg/geo — PLZ resolution (numeric and lat/lon)

**Files:**
- Create: `pkg/geo/plz.go`
- Create: `pkg/geo/plz_test.go`

- [ ] **Step 1: Write tests**

Create `pkg/geo/plz_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./pkg/geo/
```

Expected: FAIL.

- [ ] **Step 3: Write implementation**

Create `pkg/geo/plz.go`:

```go
package geo

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePLZ takes user input and returns a 6-digit PLZ code.
// Accepts: numeric PLZ (padded to 6 digits), or "lat,lon" coordinates.
// Text-based lookup is handled separately by SearchPLZ.
func ParsePLZ(input string) (string, error) {
	input = strings.TrimSpace(input)

	// Try coordinates first
	if _, _, ok := parseCoordinates(input); ok {
		// TODO: reverse geocode to nearest PLZ — for now return error
		return "", fmt.Errorf("coordinate lookup not yet supported, use a PLZ code")
	}

	// Try numeric PLZ
	if isNumeric(input) {
		return padPLZ(input), nil
	}

	return "", fmt.Errorf("location not found. Try a PLZ code (e.g. 8001) or place name")
}

func padPLZ(plz string) string {
	for len(plz) < 6 {
		plz += "0"
	}
	return plz[:6]
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseCoordinates(s string) (lat, lon float64, ok bool) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	var err error
	lat, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, false
	}
	lon, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, false
	}
	// Basic sanity check for Swiss coordinates
	if lat < 45 || lat > 48 || lon < 5 || lon > 11 {
		return lat, lon, true // valid coords but outside Switzerland — still parse
	}
	return lat, lon, true
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./pkg/geo/
```

Expected: PASS (coordinate-based PLZ test will now get an error, update the test).

Actually, update the test for ParsePLZ to handle the coordinate case:

The test `TestParsePLZ_Invalid` passes because "abc" is not numeric and not coordinates. The coordinate input `47.37,8.55` would return an error for now (TODO). Let's add a specific test:

```go
func TestParsePLZ_Coordinates_NotYetSupported(t *testing.T) {
	_, err := ParsePLZ("47.37,8.55")
	if err == nil {
		t.Error("ParsePLZ with coordinates should return error until lookup is implemented")
	}
}
```

- [ ] **Step 5: Run all tests**

```bash
go test ./pkg/geo/
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/geo/
git commit -m "feat: add geo package with PLZ parsing and coordinate detection"
```

---

## Phase 2: MeteoSwiss CLI

### Task 7: MeteoSwiss root command with global flags

**Files:**
- Modify: `cmd/meteoswiss/main.go`
- Create: `internal/meteoswiss/cmd/root.go`

- [ ] **Step 1: Add cobra dependency**

```bash
go get github.com/spf13/cobra
```

- [ ] **Step 2: Create root command**

Create `internal/meteoswiss/cmd/root.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/matthias/swisscli/pkg/config"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	version  = "dev"
	langFlag string
)

var rootCmd = &cobra.Command{
	Use:     "meteoswiss",
	Short:   "CLI for MeteoSwiss weather data",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		Lang = config.DetectLangWithEnv(langFlag, "METEOSWISS_LANG")
	},
}

var Lang string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&output.ForceJSON, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&output.NoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", "Language (de, fr, it, en)")
}
```

- [ ] **Step 3: Update entry point**

Replace `cmd/meteoswiss/main.go`:

```go
package main

import "github.com/matthias/swisscli/internal/meteoswiss/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 4: Verify it builds and runs**

```bash
go run ./cmd/meteoswiss --help
go run ./cmd/meteoswiss --version
```

Expected: help text with global flags, version output.

- [ ] **Step 5: Commit**

```bash
git add cmd/meteoswiss/ internal/meteoswiss/
git commit -m "feat: add meteoswiss root command with global flags"
```

---

### Task 8: MeteoSwiss API client

**Files:**
- Create: `internal/meteoswiss/api/client.go`
- Create: `internal/meteoswiss/api/client_test.go`

- [ ] **Step 1: Write test**

Create `internal/meteoswiss/api/client_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept header = %q, want application/json", r.Header.Get("Accept"))
		}
		if r.Header.Get("Accept-Language") != "de" {
			t.Errorf("Accept-Language = %q, want de", r.Header.Get("Accept-Language"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := &Client{http: &http.Client{}, baseURL: server.URL, lang: "de"}
	var result map[string]string
	err := c.DoJSON("GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("DoJSON error = %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %q, want ok", result["status"])
	}
}

func TestDoJSON_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := &Client{http: &http.Client{}, baseURL: server.URL, lang: "de"}
	err := c.DoJSON("GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("DoJSON should return error on 404")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/meteoswiss/api/
```

Expected: FAIL.

- [ ] **Step 3: Write implementation**

Create `internal/meteoswiss/api/client.go`:

```go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "https://app-prod-ws.meteoswiss-app.ch"
const openDataBaseURL = "https://data.geo.admin.ch"

type Client struct {
	http    *http.Client
	baseURL string
	lang    string
}

func NewClient(lang string) *Client {
	return &Client{
		http:    &http.Client{},
		baseURL: defaultBaseURL,
		lang:    lang,
	}
}

func (c *Client) DoJSON(method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", c.lang)
	req.Header.Set("User-Agent", "SwissCLI/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach API. Check your internet connection")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) DoRaw(method, url string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "SwissCLI/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach API. Check your internet connection")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/meteoswiss/api/
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/meteoswiss/api/
git commit -m "feat: add meteoswiss HTTP client with JSON and raw request methods"
```

---

### Task 9: MeteoSwiss forecast API + command

**Files:**
- Create: `internal/meteoswiss/api/forecast.go`
- Create: `internal/meteoswiss/api/forecast_test.go`
- Create: `internal/meteoswiss/cmd/forecast.go`

- [ ] **Step 1: Write API test**

Create `internal/meteoswiss/api/forecast_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetForecast(t *testing.T) {
	response := PlzDetail{
		CurrentWeather: CurrentWeather{
			Time:        "2026-04-05T18:00:00Z",
			Icon:        1,
			Temperature: 16.8,
		},
		Forecast: []ForecastDay{
			{
				DayDate:        "2026-04-05",
				IconDay:        1,
				TemperatureMax: 24,
				TemperatureMin: 10,
			},
			{
				DayDate:        "2026-04-06",
				IconDay:        2,
				TemperatureMax: 20,
				TemperatureMin: 8,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/plzDetail" {
			t.Errorf("path = %q, want /v1/plzDetail", r.URL.Path)
		}
		if r.URL.Query().Get("plz") != "800100" {
			t.Errorf("plz = %q, want 800100", r.URL.Query().Get("plz"))
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := &Client{http: &http.Client{}, baseURL: server.URL, lang: "en"}
	detail, err := c.GetForecast("800100")
	if err != nil {
		t.Fatalf("GetForecast error = %v", err)
	}
	if detail.CurrentWeather.Temperature != 16.8 {
		t.Errorf("temperature = %f, want 16.8", detail.CurrentWeather.Temperature)
	}
	if len(detail.Forecast) != 2 {
		t.Errorf("forecast days = %d, want 2", len(detail.Forecast))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/meteoswiss/api/ -run TestGetForecast
```

Expected: FAIL.

- [ ] **Step 3: Write forecast API**

Create `internal/meteoswiss/api/forecast.go`:

```go
package api

import "fmt"

type PlzDetail struct {
	CurrentWeather CurrentWeather `json:"currentWeather"`
	Forecast       []ForecastDay  `json:"forecast"`
	Warnings       []Warning      `json:"warnings,omitempty"`
	Graph          *Graph         `json:"graph,omitempty"`
}

type CurrentWeather struct {
	Time        string  `json:"time"`
	Icon        int     `json:"icon"`
	IconV2      int     `json:"iconV2"`
	Temperature float64 `json:"temperature"`
}

type ForecastDay struct {
	DayDate          string  `json:"dayDate"`
	IconDay          int     `json:"iconDay"`
	IconDayV2        int     `json:"iconDayV2"`
	TemperatureMax   float64 `json:"temperatureMax"`
	TemperatureMin   float64 `json:"temperatureMin"`
	Precipitation    float64 `json:"precipitation"`
	PrecipitationMin float64 `json:"precipitationMin"`
	PrecipitationMax float64 `json:"precipitationMax"`
}

type Warning struct {
	Type      int    `json:"warnType"`
	Level     int    `json:"warnLevel"`
	Text      string `json:"text"`
	ValidFrom string `json:"validFrom"`
	ValidTo   string `json:"validTo"`
}

type Graph struct {
	Start                string    `json:"start"`
	StartLowResolution   string    `json:"startLowResolution"`
	Precipitation10m     []float64 `json:"precipitation10m,omitempty"`
	Precipitation1h      []float64 `json:"precipitation1h,omitempty"`
	TemperatureMean1h    []float64 `json:"temperatureMean1h,omitempty"`
	TemperatureMin1h     []float64 `json:"temperatureMin1h,omitempty"`
	TemperatureMax1h     []float64 `json:"temperatureMax1h,omitempty"`
	WindSpeed1h          []float64 `json:"windSpeed1h,omitempty"`
	WindGust1h           []float64 `json:"windGust1h,omitempty"`
	WindDirection1h      []int     `json:"windDirection1h,omitempty"`
	Sunrise              []string  `json:"sunrise,omitempty"`
	Sunset               []string  `json:"sunset,omitempty"`
}

func (c *Client) GetForecast(plz string) (*PlzDetail, error) {
	var detail PlzDetail
	err := c.DoJSON("GET", fmt.Sprintf("/v1/plzDetail?plz=%s", plz), nil, &detail)
	if err != nil {
		return nil, fmt.Errorf("get forecast: %w", err)
	}
	return &detail, nil
}
```

- [ ] **Step 4: Run test**

```bash
go test ./internal/meteoswiss/api/ -run TestGetForecast
```

Expected: PASS.

- [ ] **Step 5: Write forecast command**

Create `internal/meteoswiss/cmd/forecast.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/matthias/swisscli/internal/meteoswiss/api"
	"github.com/matthias/swisscli/pkg/geo"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var weekFlag bool

func init() {
	rootCmd.AddCommand(forecastCmd)
	forecastCmd.Flags().BoolVar(&weekFlag, "week", false, "Show 8-day forecast")
}

var forecastCmd = &cobra.Command{
	Use:   "forecast <location>",
	Short: "Weather forecast for a location",
	Long:  "Show weather forecast by PLZ code (e.g. 8001), place name, or lat,lon coordinates.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		plz, err := geo.ParsePLZ(args[0])
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		client := api.NewClient(Lang)
		detail, err := client.GetForecast(plz)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{
				"currentWeather": detail.CurrentWeather,
				"forecast":       detail.Forecast,
				"warnings":       detail.Warnings,
				"source":         source.MeteoSwiss,
			})
			return nil
		}

		// Current weather
		output.Section("Current Weather")
		cw := detail.CurrentWeather
		fmt.Printf("  %s  %.1f°C  (Icon: %d)\n", cw.Time, cw.Temperature, cw.Icon)

		// Forecast table
		if weekFlag {
			output.Section("8-Day Forecast")
		} else {
			output.Section("Forecast")
		}

		days := detail.Forecast
		if !weekFlag && len(days) > 3 {
			days = days[:3]
		}

		headers := []string{"DATE", "ICON", "MIN", "MAX", "PRECIP"}
		var rows [][]string
		for _, d := range days {
			rows = append(rows, []string{
				d.DayDate,
				fmt.Sprintf("%d", d.IconDay),
				fmt.Sprintf("%.0f°C", d.TemperatureMin),
				fmt.Sprintf("%.0f°C", d.TemperatureMax),
				fmt.Sprintf("%.1f mm", d.Precipitation),
			})
		}
		output.Table(headers, rows)

		// Warnings
		if len(detail.Warnings) > 0 {
			output.Section("Warnings")
			for _, w := range detail.Warnings {
				fmt.Printf("  [Level %d] %s (%s — %s)\n", w.Level, w.Text, w.ValidFrom, w.ValidTo)
			}
		}

		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
```

- [ ] **Step 6: Verify it builds**

```bash
go build ./cmd/meteoswiss
```

Expected: builds successfully.

- [ ] **Step 7: Test with real API**

```bash
go run ./cmd/meteoswiss forecast 8001
go run ./cmd/meteoswiss forecast 8001 --week
go run ./cmd/meteoswiss forecast 8001 --json
```

Expected: shows current weather + forecast table / 8-day table / JSON output.

- [ ] **Step 8: Commit**

```bash
git add internal/meteoswiss/
git commit -m "feat: add meteoswiss forecast command with plzDetail API"
```

---

### Task 10: MeteoSwiss current measurements command

**Files:**
- Create: `internal/meteoswiss/api/measurements.go`
- Create: `internal/meteoswiss/api/measurements_test.go`
- Create: `internal/meteoswiss/cmd/current.go`

- [ ] **Step 1: Write API test**

Create `internal/meteoswiss/api/measurements_test.go`:

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseVQHA80(t *testing.T) {
	csv := "Station/Location;Date;tre200s0;rre150z0;sre000z0;gre000z0;ure200s0;tde200s0;dkl010z0;fu3010z0;fu3010z1;prestas0;pp0qffs0;pp0qnhs0;ppz850s0;ppz700s0;dv1towz0;fu3towz0;fu3towz1;ta1tows0;uretows0;tdetows0\nTAE;202604051910;15.2;0.0;10;245;55;6.3;250;12.6;22.3;965.1;1015.2;1015.0;-;-;-;-;-;-;-;-\nSMA;202604051910;18.1;0.0;10;180;42;5.1;210;8.4;18.7;955.8;1013.8;1013.6;-;-;230;14.0;28.1;17.2;40;3.6\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(csv))
	}))
	defer server.Close()

	c := &Client{http: &http.Client{}, baseURL: server.URL, lang: "de"}
	stations, err := c.GetCurrentMeasurements(server.URL + "/test.csv")
	if err != nil {
		t.Fatalf("GetCurrentMeasurements error = %v", err)
	}
	if len(stations) != 2 {
		t.Fatalf("stations = %d, want 2", len(stations))
	}
	if stations[0].Station != "TAE" {
		t.Errorf("station[0] = %q, want TAE", stations[0].Station)
	}
	if stations[0].Temperature != "15.2" {
		t.Errorf("temperature = %q, want 15.2", stations[0].Temperature)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/meteoswiss/api/ -run TestParseVQHA80
```

Expected: FAIL.

- [ ] **Step 3: Write measurements API**

Create `internal/meteoswiss/api/measurements.go`:

```go
package api

import (
	"encoding/csv"
	"fmt"
	"strings"
)

const measurementsURL = "https://data.geo.admin.ch/ch.meteoschweiz.messwerte-aktuell/VQHA80.csv"

type StationMeasurement struct {
	Station     string `json:"station"`
	Date        string `json:"date"`
	Temperature string `json:"temperature"`
	Rainfall    string `json:"rainfall"`
	Sunshine    string `json:"sunshine"`
	Radiation   string `json:"radiation"`
	Humidity    string `json:"humidity"`
	DewPoint    string `json:"dew_point"`
	WindDir     string `json:"wind_direction"`
	WindSpeed   string `json:"wind_speed"`
	GustPeak    string `json:"gust_peak"`
	Pressure    string `json:"pressure_station"`
	PressureQFE string `json:"pressure_qfe"`
	PressureQNH string `json:"pressure_qnh"`
}

func (c *Client) GetCurrentMeasurements(url string) ([]StationMeasurement, error) {
	if url == "" {
		url = measurementsURL
	}

	data, err := c.DoRaw("GET", url)
	if err != nil {
		return nil, fmt.Errorf("fetch measurements: %w", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = ';'
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("no measurement data found")
	}

	var measurements []StationMeasurement
	for _, row := range records[1:] {
		if len(row) < 14 {
			continue
		}
		m := StationMeasurement{
			Station:     row[0],
			Date:        row[1],
			Temperature: dashToEmpty(row[2]),
			Rainfall:    dashToEmpty(row[3]),
			Sunshine:    dashToEmpty(row[4]),
			Radiation:   dashToEmpty(row[5]),
			Humidity:    dashToEmpty(row[6]),
			DewPoint:    dashToEmpty(row[7]),
			WindDir:     dashToEmpty(row[8]),
			WindSpeed:   dashToEmpty(row[9]),
			GustPeak:    dashToEmpty(row[10]),
			Pressure:    dashToEmpty(row[11]),
			PressureQFE: dashToEmpty(row[12]),
			PressureQNH: dashToEmpty(row[13]),
		}
		measurements = append(measurements, m)
	}
	return measurements, nil
}

func dashToEmpty(s string) string {
	if s == "-" {
		return ""
	}
	return s
}
```

- [ ] **Step 4: Run test**

```bash
go test ./internal/meteoswiss/api/ -run TestParseVQHA80
```

Expected: PASS.

- [ ] **Step 5: Write current command**

Create `internal/meteoswiss/cmd/current.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/matthias/swisscli/internal/meteoswiss/api"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(currentCmd)
}

var currentCmd = &cobra.Command{
	Use:   "current [station]",
	Short: "Current weather measurements",
	Long:  "Show current measurements from automatic weather stations. Optionally filter by station code (e.g. SMA for Zürich).",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)
		measurements, err := client.GetCurrentMeasurements("")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		// Filter by station if provided
		if len(args) > 0 {
			search := strings.ToUpper(args[0])
			var filtered []api.StationMeasurement
			for _, m := range measurements {
				if m.Station == search {
					filtered = append(filtered, m)
				}
			}
			if len(filtered) == 0 {
				output.Error(fmt.Sprintf("station %q not found", args[0]))
				os.Exit(1)
			}
			measurements = filtered
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{
				"measurements": measurements,
				"source":       source.MeteoSwiss,
			})
			return nil
		}

		output.Section("Current Measurements")
		headers := []string{"STATION", "TEMP", "HUMIDITY", "WIND", "GUSTS", "PRESSURE", "RAIN"}
		var rows [][]string
		for _, m := range measurements {
			rows = append(rows, []string{
				m.Station,
				fmtVal(m.Temperature, "°C"),
				fmtVal(m.Humidity, "%"),
				fmtVal(m.WindSpeed, " km/h"),
				fmtVal(m.GustPeak, " km/h"),
				fmtVal(m.PressureQNH, " hPa"),
				fmtVal(m.Rainfall, " mm"),
			})
		}
		output.Table(headers, rows)
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}

func fmtVal(val, unit string) string {
	if val == "" {
		return "-"
	}
	return val + unit
}
```

- [ ] **Step 6: Verify it builds and test with real API**

```bash
go run ./cmd/meteoswiss current
go run ./cmd/meteoswiss current SMA
go run ./cmd/meteoswiss current SMA --json
```

Expected: table of all stations / single station / JSON output.

- [ ] **Step 7: Commit**

```bash
git add internal/meteoswiss/
git commit -m "feat: add meteoswiss current measurements command (VQHA80 CSV)"
```

---

### Task 11: MeteoSwiss radar command

**Files:**
- Create: `internal/meteoswiss/api/radar.go`
- Create: `internal/meteoswiss/cmd/radar.go`

- [ ] **Step 1: Write radar API with URL mappings**

Create `internal/meteoswiss/api/radar.go`:

```go
package api

type RadarType string

const (
	RadarRain      RadarType = "rain"
	RadarCloud     RadarType = "cloud"
	RadarSatellite RadarType = "satellite"
)

var radarBrowserURLs = map[RadarType]string{
	RadarRain:      "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/niederschlag.html",
	RadarCloud:     "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/satellitenbilder.html#tab=satellite-animation-hrv",
	RadarSatellite: "https://www.meteoschweiz.admin.ch/service-und-publikationen/applikationen/satellitenbilder.html#tab=satellite-animation-hrv",
}

var radarImageURLs = map[RadarType]string{
	RadarRain:      "https://www.meteoschweiz.admin.ch/product/output/radar/precip/animation/radar_precip.png",
	RadarCloud:     "https://www.meteoschweiz.admin.ch/product/output/satellite/animation/satellite_hrv.png",
	RadarSatellite: "https://www.meteoschweiz.admin.ch/product/output/satellite/animation/satellite_hrv.png",
}

func GetRadarBrowserURL(rt RadarType) string {
	return radarBrowserURLs[rt]
}

func GetRadarImageURL(rt RadarType) string {
	return radarImageURLs[rt]
}
```

- [ ] **Step 2: Write radar command**

Create `internal/meteoswiss/cmd/radar.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matthias/swisscli/internal/meteoswiss/api"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	radarSave  string
	radarASCII bool
	radarWidth int
)

func init() {
	rootCmd.AddCommand(radarCmd)
	radarCmd.Flags().StringVar(&radarSave, "save", "", "Save image to file path")
	radarCmd.Flags().BoolVar(&radarASCII, "ascii", false, "Render as ASCII art in terminal")
	radarCmd.Flags().IntVar(&radarWidth, "width", 80, "ASCII art width in columns")
}

var radarCmd = &cobra.Command{
	Use:   "radar [rain|cloud|satellite]",
	Short: "Weather radar and satellite images",
	Long:  "View rain radar, cloud radar, or satellite imagery. Default: opens in browser.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		radarType := api.RadarRain
		if len(args) > 0 {
			switch args[0] {
			case "rain":
				radarType = api.RadarRain
			case "cloud":
				radarType = api.RadarCloud
			case "satellite":
				radarType = api.RadarSatellite
			default:
				output.Error(fmt.Sprintf("unknown radar type %q — use rain, cloud, or satellite", args[0]))
				os.Exit(1)
			}
		}

		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"type":       string(radarType),
				"browser_url": api.GetRadarBrowserURL(radarType),
				"image_url":   api.GetRadarImageURL(radarType),
				"source":      source.MeteoSwiss,
			})
			return nil
		}

		if radarASCII {
			fmt.Printf("Fetching %s radar...\n", radarType)
			if err := output.ASCIIMap(api.GetRadarImageURL(radarType), radarWidth); err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			fmt.Printf("\n%s\n", source.MeteoSwiss)
			return nil
		}

		if radarSave != "" {
			path := radarSave
			if filepath.Ext(path) == "" {
				path += ".png"
			}
			fmt.Printf("Saving %s radar to %s...\n", radarType, path)
			if err := output.SaveImage(api.GetRadarImageURL(radarType), path); err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			fmt.Printf("Saved to %s\n", path)
			fmt.Printf("\n%s\n", source.MeteoSwiss)
			return nil
		}

		// Default: open in browser
		url := api.GetRadarBrowserURL(radarType)
		fmt.Printf("Opening %s radar in browser...\n", radarType)
		if err := output.OpenBrowser(url); err != nil {
			// Fallback: print URL
			fmt.Printf("Could not open browser. Visit: %s\n", url)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
```

- [ ] **Step 3: Verify it builds**

```bash
go build ./cmd/meteoswiss
```

- [ ] **Step 4: Test interactively**

```bash
go run ./cmd/meteoswiss radar rain
go run ./cmd/meteoswiss radar rain --ascii
go run ./cmd/meteoswiss radar --json
```

Expected: opens browser / ASCII art / JSON with URLs.

- [ ] **Step 5: Commit**

```bash
git add internal/meteoswiss/
git commit -m "feat: add meteoswiss radar command with browser, save, and ASCII modes"
```

---

### Task 12: MeteoSwiss stations command

**Files:**
- Create: `internal/meteoswiss/cmd/stations.go`

- [ ] **Step 1: Write stations command**

Create `internal/meteoswiss/cmd/stations.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/matthias/swisscli/internal/meteoswiss/api"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var stationsSearch string

func init() {
	rootCmd.AddCommand(stationsCmd)
	stationsCmd.Flags().StringVar(&stationsSearch, "search", "", "Search stations by code")
}

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "List weather measurement stations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)
		measurements, err := client.GetCurrentMeasurements("")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		// Deduplicate station codes
		seen := make(map[string]bool)
		var stations []api.StationMeasurement
		for _, m := range measurements {
			if seen[m.Station] {
				continue
			}
			seen[m.Station] = true
			if stationsSearch != "" && !strings.Contains(strings.ToUpper(m.Station), strings.ToUpper(stationsSearch)) {
				continue
			}
			stations = append(stations, m)
		}

		if !output.IsInteractive() {
			codes := make([]string, len(stations))
			for i, s := range stations {
				codes[i] = s.Station
			}
			output.JSON(map[string]any{
				"stations": codes,
				"count":    len(codes),
				"source":   source.MeteoSwiss,
			})
			return nil
		}

		output.Section(fmt.Sprintf("Stations (%d)", len(stations)))
		headers := []string{"CODE", "TEMP", "WIND", "RAIN"}
		var rows [][]string
		for _, s := range stations {
			rows = append(rows, []string{
				s.Station,
				fmtVal(s.Temperature, "°C"),
				fmtVal(s.WindSpeed, " km/h"),
				fmtVal(s.Rainfall, " mm"),
			})
		}
		output.Table(headers, rows)
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
```

- [ ] **Step 2: Verify and test**

```bash
go run ./cmd/meteoswiss stations
go run ./cmd/meteoswiss stations --search SMA
go run ./cmd/meteoswiss stations --json
```

- [ ] **Step 3: Commit**

```bash
git add internal/meteoswiss/cmd/stations.go
git commit -m "feat: add meteoswiss stations command"
```

---

### Task 13: MeteoSwiss favorites command

**Files:**
- Create: `internal/meteoswiss/cmd/favorites.go`

- [ ] **Step 1: Write favorites command**

Create `internal/meteoswiss/cmd/favorites.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/matthias/swisscli/pkg/config"
	"github.com/matthias/swisscli/pkg/geo"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(favoritesCmd)
	favoritesCmd.AddCommand(favoritesAddCmd)
	favoritesCmd.AddCommand(favoritesRemoveCmd)
	favoritesCmd.AddCommand(favoritesListCmd)
}

var favoritesCmd = &cobra.Command{
	Use:   "favorites",
	Short: "Manage saved locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return favoritesListCmd.RunE(cmd, args)
	},
}

var favoritesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("meteoswiss")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if !output.IsInteractive() {
			output.JSON(cfg.Favorites)
			return nil
		}

		if len(cfg.Favorites) == 0 {
			fmt.Println("No favorites saved. Use `meteoswiss favorites add <name> <plz>` to add one.")
			return nil
		}

		headers := []string{"NAME", "PLZ", "STATION"}
		var rows [][]string
		for _, f := range cfg.Favorites {
			rows = append(rows, []string{f.Name, f.PLZ, f.Station})
		}
		output.Table(headers, rows)
		return nil
	},
}

var favoritesAddCmd = &cobra.Command{
	Use:   "add <name> <location>",
	Short: "Add a favorite location",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		location := args[1]

		plz, err := geo.ParsePLZ(location)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		cfg, err := config.LoadDefault("meteoswiss")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		cfg.AddFavorite(config.Favorite{Name: name, PLZ: plz})
		if err := cfg.Save(); err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Added %q (PLZ: %s) to favorites.\n", name, plz)
		return nil
	},
}

var favoritesRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a favorite location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("meteoswiss")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		cfg.RemoveFavorite(args[0])
		if err := cfg.Save(); err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Removed %q from favorites.\n", args[0])
		return nil
	},
}
```

- [ ] **Step 2: Verify and test**

```bash
go run ./cmd/meteoswiss favorites list
go run ./cmd/meteoswiss favorites add Zürich 8001
go run ./cmd/meteoswiss favorites list
go run ./cmd/meteoswiss favorites remove Zürich
```

- [ ] **Step 3: Commit**

```bash
git add internal/meteoswiss/cmd/favorites.go
git commit -m "feat: add meteoswiss favorites command"
```

---

### Task 14: MeteoSwiss hazards, wind, pollen, precipitation, clouds commands

These commands follow a similar pattern — some are data-based (wind, pollen, precipitation from plzDetail graph data or open data CSVs), others are primarily browser-based (hazards map, clouds visualization). We'll implement them as a batch since they follow the same structure.

**Files:**
- Create: `internal/meteoswiss/cmd/hazards.go`
- Create: `internal/meteoswiss/cmd/wind.go`
- Create: `internal/meteoswiss/cmd/pollen.go`
- Create: `internal/meteoswiss/cmd/precipitation.go`
- Create: `internal/meteoswiss/cmd/clouds.go`

- [ ] **Step 1: Write hazards command**

Create `internal/meteoswiss/cmd/hazards.go`:

```go
package cmd

import (
	"fmt"

	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

const hazardsURL = "https://www.meteoswiss.admin.ch/services-and-publications/applications/hazards.html#tab=natural-hazards-map"

func init() {
	rootCmd.AddCommand(hazardsCmd)
	hazardsCmd.Flags().BoolVar(&radarASCII, "ascii", false, "ASCII art mode")
	hazardsCmd.Flags().StringVar(&radarSave, "save", "", "Save image to file")
}

var hazardsCmd = &cobra.Command{
	Use:   "hazards",
	Short: "Natural hazard warnings",
	Long:  "View the current natural hazards map for Switzerland.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"url":    hazardsURL,
				"source": source.MeteoSwiss,
			})
			return nil
		}

		fmt.Println("Opening natural hazards map in browser...")
		if err := output.OpenBrowser(hazardsURL); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", hazardsURL)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
```

- [ ] **Step 2: Write wind command**

Create `internal/meteoswiss/cmd/wind.go`:

```go
package cmd

import (
	"fmt"

	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var windLevel string

const windURL = "https://www.meteoswiss.admin.ch/services-and-publications/applications/wind.html#tab=animation-wind-10m"

func init() {
	rootCmd.AddCommand(windCmd)
	windCmd.Flags().StringVar(&windLevel, "level", "10m", "Wind level: 10m, 2000m, gust")
}

var windCmd = &cobra.Command{
	Use:   "wind",
	Short: "Wind data and animations",
	Long:  "View wind data at 10m, 2000m (upper winds), or gust peaks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		urlMap := map[string]string{
			"10m":  "https://www.meteoswiss.admin.ch/services-and-publications/applications/wind.html#tab=animation-wind-10m",
			"2000m": "https://www.meteoswiss.admin.ch/services-and-publications/applications/wind.html#tab=animation-upper-winds-2000m",
			"gust": "https://www.meteoswiss.admin.ch/services-and-publications/applications/wind.html#tab=animation-gust-peaks-10m",
		}

		url, ok := urlMap[windLevel]
		if !ok {
			output.Error(fmt.Sprintf("unknown wind level %q — use 10m, 2000m, or gust", windLevel))
			return nil
		}

		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"level":  windLevel,
				"url":    url,
				"source": source.MeteoSwiss,
			})
			return nil
		}

		fmt.Printf("Opening wind (%s) in browser...\n", windLevel)
		if err := output.OpenBrowser(url); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", url)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
```

- [ ] **Step 3: Write pollen command**

Create `internal/meteoswiss/cmd/pollen.go`:

```go
package cmd

import (
	"fmt"

	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var pollenType string

func init() {
	rootCmd.AddCommand(pollenCmd)
	pollenCmd.Flags().StringVar(&pollenType, "type", "all", "Pollen type filter")
}

var pollenCmd = &cobra.Command{
	Use:   "pollen",
	Short: "Pollen forecast and report",
	RunE: func(cmd *cobra.Command, args []string) error {
		mapURL := fmt.Sprintf("https://www.meteoswiss.admin.ch/services-and-publications/applications/pollen-forecast.html#tab=pollen-map&pollen=%s", pollenType)
		forecastURL := fmt.Sprintf("https://www.meteoswiss.admin.ch/services-and-publications/applications/pollen-forecast.html#tab=pollen-forecast&pollen=%s", pollenType)

		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"map_url":      mapURL,
				"forecast_url": forecastURL,
				"pollen_type":  pollenType,
				"source":       source.MeteoSwiss,
			})
			return nil
		}

		fmt.Println("Opening pollen forecast in browser...")
		if err := output.OpenBrowser(forecastURL); err != nil {
			fmt.Printf("Could not open browser. Visit:\n  Map: %s\n  Forecast: %s\n", mapURL, forecastURL)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
```

- [ ] **Step 4: Write precipitation command**

Create `internal/meteoswiss/cmd/precipitation.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/matthias/swisscli/internal/meteoswiss/api"
	"github.com/matthias/swisscli/pkg/geo"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(precipitationCmd)
}

var precipitationCmd = &cobra.Command{
	Use:   "precipitation <location>",
	Short: "Precipitation probability",
	Long:  "Show precipitation data from the forecast graph for a given location.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		plz, err := geo.ParsePLZ(args[0])
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		client := api.NewClient(Lang)
		detail, err := client.GetForecast(plz)
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{
				"forecast":      detail.Forecast,
				"graph":         detail.Graph,
				"source":        source.MeteoSwiss,
			})
			return nil
		}

		output.Section("Precipitation Forecast")
		headers := []string{"DATE", "PRECIP", "MIN", "MAX"}
		var rows [][]string
		for _, d := range detail.Forecast {
			rows = append(rows, []string{
				d.DayDate,
				fmt.Sprintf("%.1f mm", d.Precipitation),
				fmt.Sprintf("%.1f mm", d.PrecipitationMin),
				fmt.Sprintf("%.1f mm", d.PrecipitationMax),
			})
		}
		output.Table(headers, rows)
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
```

- [ ] **Step 5: Write clouds command**

Create `internal/meteoswiss/cmd/clouds.go`:

```go
package cmd

import (
	"fmt"

	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

const cloudsURL = "https://www.meteoswiss.admin.ch/services-and-publications/applications/cloud-cover.html#tab=cloud-coverage"

func init() {
	rootCmd.AddCommand(cloudsCmd)
}

var cloudsCmd = &cobra.Command{
	Use:   "clouds",
	Short: "Cloud cover visualization",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"url":    cloudsURL,
				"source": source.MeteoSwiss,
			})
			return nil
		}

		fmt.Println("Opening cloud cover in browser...")
		if err := output.OpenBrowser(cloudsURL); err != nil {
			fmt.Printf("Could not open browser. Visit: %s\n", cloudsURL)
		}
		fmt.Printf("\n%s\n", source.MeteoSwiss)
		return nil
	},
}
```

- [ ] **Step 6: Verify all commands build**

```bash
go build ./cmd/meteoswiss
./meteoswiss --help
```

Expected: all commands listed in help.

- [ ] **Step 7: Commit**

```bash
git add internal/meteoswiss/cmd/
git commit -m "feat: add hazards, wind, pollen, precipitation, and clouds commands"
```

---

## Phase 3: WhiteRisk CLI

### Task 15: WhiteRisk root command

**Files:**
- Create: `internal/whiterisk/cmd/root.go`
- Modify: `cmd/whiterisk/main.go`

- [ ] **Step 1: Create root command**

Create `internal/whiterisk/cmd/root.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/matthias/swisscli/pkg/config"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	version  = "dev"
	langFlag string
)

var rootCmd = &cobra.Command{
	Use:     "whiterisk",
	Short:   "CLI for SLF/WSL avalanche and snow data",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		Lang = config.DetectLangWithEnv(langFlag, "WHITERISK_LANG")
	},
}

var Lang string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&output.ForceJSON, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&output.NoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", "Language (de, fr, it, en)")
}
```

- [ ] **Step 2: Update entry point**

Replace `cmd/whiterisk/main.go`:

```go
package main

import "github.com/matthias/swisscli/internal/whiterisk/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 3: Verify**

```bash
go run ./cmd/whiterisk --help
```

- [ ] **Step 4: Commit**

```bash
git add cmd/whiterisk/ internal/whiterisk/
git commit -m "feat: add whiterisk root command with global flags"
```

---

### Task 16: WhiteRisk API client

**Files:**
- Create: `internal/whiterisk/api/client.go`
- Create: `internal/whiterisk/api/client_test.go`

- [ ] **Step 1: Write test**

Create `internal/whiterisk/api/client_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClientWithBase(server.URL, "en")
	var result map[string]string
	err := c.DoJSON("GET", "/test", &result)
	if err != nil {
		t.Fatalf("DoJSON error = %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %q, want ok", result["status"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/whiterisk/api/
```

- [ ] **Step 3: Write implementation**

Create `internal/whiterisk/api/client.go`:

```go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	bulletinBaseURL     = "https://aws.slf.ch"
	measurementBaseURL  = "https://measurement-api.slf.ch"
	whiteriskBaseURL    = "https://whiterisk.ch"
)

type Client struct {
	http           *http.Client
	bulletinBase   string
	measurementBase string
	lang           string
}

func NewClient(lang string) *Client {
	return &Client{
		http:           &http.Client{},
		bulletinBase:   bulletinBaseURL,
		measurementBase: measurementBaseURL,
		lang:           lang,
	}
}

func NewClientWithBase(baseURL, lang string) *Client {
	return &Client{
		http:           &http.Client{},
		bulletinBase:   baseURL,
		measurementBase: baseURL,
		lang:           lang,
	}
}

func (c *Client) DoJSON(method, url string, result any) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SwissCLI/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach API. Check your internet connection")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) DoRaw(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "SwissCLI/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach API. Check your internet connection")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
```

- [ ] **Step 4: Run test**

```bash
go test ./internal/whiterisk/api/
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/whiterisk/api/
git commit -m "feat: add whiterisk HTTP client"
```

---

### Task 17: WhiteRisk bulletin API + command

**Files:**
- Create: `internal/whiterisk/api/bulletin.go`
- Create: `internal/whiterisk/api/bulletin_test.go`
- Create: `internal/whiterisk/cmd/bulletin.go`

- [ ] **Step 1: Write bulletin API test**

Create `internal/whiterisk/api/bulletin_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/whiterisk/api/ -run TestGetBulletin
```

- [ ] **Step 3: Write bulletin API**

Create `internal/whiterisk/api/bulletin.go`:

```go
package api

import "fmt"

type BulletinResponse struct {
	Bulletins  []Bulletin `json:"bulletins"`
	CustomData any        `json:"customData,omitempty"`
}

type Bulletin struct {
	BulletinID      string         `json:"bulletinID"`
	Lang            string         `json:"lang"`
	PublicationTime string         `json:"publicationTime"`
	ValidTime       ValidTime      `json:"validTime"`
	NextUpdate      string         `json:"nextUpdate,omitempty"`
	Unscheduled     bool           `json:"unscheduled,omitempty"`
	Regions         []Region       `json:"regions"`
	DangerRatings   []DangerRating `json:"dangerRatings,omitempty"`
	AvalancheProblems []AvalancheProblem `json:"avalancheProblems,omitempty"`
	WeatherForecast *TextContent   `json:"weatherForecast,omitempty"`
	SnowpackStructure *TextContent `json:"snowpackStructure,omitempty"`
	TravelAdvisory  *TextContent   `json:"travelAdvisory,omitempty"`
	Tendency        []Tendency     `json:"tendency,omitempty"`
}

type ValidTime struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type Region struct {
	RegionID string `json:"regionID"`
	Name     string `json:"name"`
}

type DangerRating struct {
	MainValue string         `json:"mainValue"`
	Elevation ElevationRange `json:"elevation,omitempty"`
	ValidTimePeriod string   `json:"validTimePeriod,omitempty"`
}

type ElevationRange struct {
	LowerBound string `json:"lowerBound,omitempty"`
	UpperBound string `json:"upperBound,omitempty"`
}

type AvalancheProblem struct {
	ProblemType string         `json:"problemType"`
	Elevation   ElevationRange `json:"elevation,omitempty"`
	Aspects     []string       `json:"aspects,omitempty"`
	ValidTimePeriod string     `json:"validTimePeriod,omitempty"`
}

type TextContent struct {
	Comment string `json:"comment,omitempty"`
}

type Tendency struct {
	Comment       string `json:"comment,omitempty"`
	TendencyType  string `json:"tendencyType,omitempty"`
	ValidTime     *ValidTime `json:"validTime,omitempty"`
}

func (c *Client) GetBulletin() (*BulletinResponse, error) {
	url := fmt.Sprintf("%s/api/bulletin/caaml/%s/json", c.bulletinBase, c.lang)
	var result BulletinResponse
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get bulletin: %w", err)
	}
	return &result, nil
}

func (c *Client) GetBulletinPDFURL() string {
	return fmt.Sprintf("%s/api/bulletin/document/full/%s", c.bulletinBase, c.lang)
}

var dangerLevelNames = map[string]string{
	"low":          "1 — Low",
	"moderate":     "2 — Moderate",
	"considerable": "3 — Considerable",
	"high":         "4 — High",
	"very_high":    "5 — Very High",
}

func DangerLevelDisplay(value string) string {
	if name, ok := dangerLevelNames[value]; ok {
		return name
	}
	return value
}
```

- [ ] **Step 4: Run test**

```bash
go test ./internal/whiterisk/api/ -run TestGetBulletin
```

Expected: PASS.

- [ ] **Step 5: Write bulletin command**

Create `internal/whiterisk/cmd/bulletin.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/matthias/swisscli/internal/whiterisk/api"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var bulletinPDF bool

func init() {
	rootCmd.AddCommand(bulletinCmd)
	bulletinCmd.Flags().BoolVar(&bulletinPDF, "pdf", false, "Download bulletin as PDF")
}

var bulletinCmd = &cobra.Command{
	Use:   "bulletin [location]",
	Short: "Avalanche bulletin",
	Long:  "Show avalanche danger ratings. Filter by region name, region ID, or lat,lon.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)

		if bulletinPDF {
			url := client.GetBulletinPDFURL()
			if !output.IsInteractive() {
				output.JSON(map[string]string{"pdf_url": url, "source": source.SLF})
				return nil
			}
			fmt.Println("Opening bulletin PDF in browser...")
			if err := output.OpenBrowser(url); err != nil {
				fmt.Printf("Download: %s\n", url)
			}
			return nil
		}

		result, err := client.GetBulletin()
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		// Filter bulletins by location if provided
		bulletins := result.Bulletins
		if len(args) > 0 {
			search := args[0]
			bulletins = filterBulletins(bulletins, search)
			if len(bulletins) == 0 {
				output.Error(fmt.Sprintf("no bulletin found for %q", search))
				os.Exit(1)
			}
		}

		if !output.IsInteractive() {
			output.JSON(map[string]any{
				"bulletins": bulletins,
				"source":    source.SLF,
			})
			return nil
		}

		for _, b := range bulletins {
			regions := make([]string, len(b.Regions))
			for i, r := range b.Regions {
				regions[i] = fmt.Sprintf("%s (%s)", r.Name, r.RegionID)
			}

			output.Section("Avalanche Bulletin")
			fmt.Printf("  Regions: %s\n", strings.Join(regions, ", "))
			fmt.Printf("  Valid:   %s → %s\n", b.ValidTime.StartTime, b.ValidTime.EndTime)

			if len(b.DangerRatings) > 0 {
				for _, dr := range b.DangerRatings {
					elev := ""
					if dr.Elevation.UpperBound != "" {
						elev = fmt.Sprintf(" (below %s)", dr.Elevation.UpperBound)
					}
					if dr.Elevation.LowerBound != "" {
						elev = fmt.Sprintf(" (above %s)", dr.Elevation.LowerBound)
					}
					fmt.Printf("  Danger:  %s%s\n", api.DangerLevelDisplay(dr.MainValue), elev)
				}
			}

			if len(b.AvalancheProblems) > 0 {
				fmt.Println("  Problems:")
				for _, p := range b.AvalancheProblems {
					aspects := strings.Join(p.Aspects, "/")
					elev := ""
					if p.Elevation.LowerBound != "" {
						elev = fmt.Sprintf(" above %s", p.Elevation.LowerBound)
					}
					fmt.Printf("    - %s — %s%s\n", p.ProblemType, aspects, elev)
				}
			}

			if b.SnowpackStructure != nil && b.SnowpackStructure.Comment != "" {
				fmt.Printf("  Snowpack: %s\n", b.SnowpackStructure.Comment)
			}
		}
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	},
}

func filterBulletins(bulletins []api.Bulletin, search string) []api.Bulletin {
	search = strings.ToLower(search)
	var matched []api.Bulletin
	for _, b := range bulletins {
		for _, r := range b.Regions {
			if strings.ToLower(r.RegionID) == strings.ToLower(search) ||
				strings.Contains(strings.ToLower(r.Name), search) {
				matched = append(matched, b)
				break
			}
		}
	}
	return matched
}
```

- [ ] **Step 6: Verify and test**

```bash
go run ./cmd/whiterisk bulletin
go run ./cmd/whiterisk bulletin Davos
go run ./cmd/whiterisk bulletin --json
go run ./cmd/whiterisk bulletin --pdf
```

- [ ] **Step 7: Commit**

```bash
git add internal/whiterisk/
git commit -m "feat: add whiterisk bulletin command with region filtering"
```

---

### Task 18: WhiteRisk stations + measurements commands

**Files:**
- Create: `internal/whiterisk/api/measurements.go`
- Create: `internal/whiterisk/api/measurements_test.go`
- Create: `internal/whiterisk/cmd/stations.go`
- Create: `internal/whiterisk/cmd/measurements.go`

- [ ] **Step 1: Write API test**

Create `internal/whiterisk/api/measurements_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIMISStations(t *testing.T) {
	stations := []IMISStation{
		{Code: "DAV2", Label: "Davos", Lon: 9.85, Lat: 46.81, Elevation: 2560, CountryCode: "CH", CantonCode: "GR", Type: "SNOW_FLAT"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(stations)
	}))
	defer server.Close()

	c := NewClientWithBase(server.URL, "en")
	result, err := c.GetIMISStations()
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(result) != 1 || result[0].Code != "DAV2" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestGetIMISMeasurements(t *testing.T) {
	measurements := []IMISMeasurement{
		{StationCode: "DAV2", MeasureDate: "2026-04-05T18:00:00Z", HS: floatPtr(142), TA30MinMean: floatPtr(-2.3)},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(measurements)
	}))
	defer server.Close()

	c := NewClientWithBase(server.URL, "en")
	result, err := c.GetIMISMeasurementsByStation("DAV2", 1)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(result) != 1 || result[0].StationCode != "DAV2" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func floatPtr(f float64) *float64 { return &f }
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/whiterisk/api/ -run "IMIS"
```

- [ ] **Step 3: Write measurements API**

Create `internal/whiterisk/api/measurements.go`:

```go
package api

import "fmt"

type IMISStation struct {
	Code        string  `json:"code"`
	Label       string  `json:"label"`
	Lon         float64 `json:"lon"`
	Lat         float64 `json:"lat"`
	Elevation   float64 `json:"elevation"`
	CountryCode string  `json:"country_code"`
	CantonCode  string  `json:"canton_code"`
	Type        string  `json:"type"`
}

type StudyPlotStation struct {
	Code        string  `json:"code"`
	Label       string  `json:"label"`
	Lon         float64 `json:"lon"`
	Lat         float64 `json:"lat"`
	Elevation   float64 `json:"elevation"`
	CountryCode string  `json:"country_code"`
	CantonCode  string  `json:"canton_code"`
}

type IMISMeasurement struct {
	StationCode  string   `json:"station_code"`
	MeasureDate  string   `json:"measure_date"`
	HS           *float64 `json:"HS"`
	TA30MinMean  *float64 `json:"TA_30MIN_MEAN"`
	RH30MinMean  *float64 `json:"RH_30MIN_MEAN"`
	TSS30MinMean *float64 `json:"TSS_30MIN_MEAN"`
	VW30MinMean  *float64 `json:"VW_30MIN_MEAN"`
	VW30MinMax   *float64 `json:"VW_30MIN_MAX"`
	DW30MinMean  *float64 `json:"DW_30MIN_MEAN"`
	RSWR30MinMean *float64 `json:"RSWR_30MIN_MEAN"`
}

type StudyPlotMeasurement struct {
	StationCode string   `json:"station_code"`
	MeasureDate string   `json:"measure_date"`
	HS          *float64 `json:"HS"`
	HN1D        *float64 `json:"HN_1D"`
	HNW1D       *float64 `json:"HNW_1D"`
}

func (c *Client) GetIMISStations() ([]IMISStation, error) {
	url := fmt.Sprintf("%s/public/api/imis/stations", c.measurementBase)
	var result []IMISStation
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get IMIS stations: %w", err)
	}
	return result, nil
}

func (c *Client) GetStudyPlotStations() ([]StudyPlotStation, error) {
	url := fmt.Sprintf("%s/public/api/study-plot/stations", c.measurementBase)
	var result []StudyPlotStation
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get study plot stations: %w", err)
	}
	return result, nil
}

func (c *Client) GetIMISMeasurementsByStation(code string, periodDays int) ([]IMISMeasurement, error) {
	url := fmt.Sprintf("%s/public/api/imis/station/%s/measurements?period_in_days=%d", c.measurementBase, code, periodDays)
	var result []IMISMeasurement
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get IMIS measurements: %w", err)
	}
	return result, nil
}

func (c *Client) GetStudyPlotMeasurementsByStation(code string, periodDays int) ([]StudyPlotMeasurement, error) {
	url := fmt.Sprintf("%s/public/api/study-plot/station/%s/measurements?period_in_days=%d", c.measurementBase, code, periodDays)
	var result []StudyPlotMeasurement
	if err := c.DoJSON("GET", url, &result); err != nil {
		return nil, fmt.Errorf("get study plot measurements: %w", err)
	}
	return result, nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/whiterisk/api/ -run "IMIS"
```

Expected: PASS.

- [ ] **Step 5: Write stations command**

Create `internal/whiterisk/cmd/stations.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/matthias/swisscli/internal/whiterisk/api"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	stationsSearch string
	stationsType   string
)

func init() {
	rootCmd.AddCommand(stationsCmd)
	stationsCmd.Flags().StringVar(&stationsSearch, "search", "", "Search by name or code")
	stationsCmd.Flags().StringVar(&stationsType, "type", "imis", "Station type: imis or study-plot")
}

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "List measurement stations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(Lang)

		if stationsType == "study-plot" {
			return listStudyPlotStations(client)
		}
		return listIMISStations(client)
	},
}

func listIMISStations(client *api.Client) error {
	stations, err := client.GetIMISStations()
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if stationsSearch != "" {
		search := strings.ToUpper(stationsSearch)
		var filtered []api.IMISStation
		for _, s := range stations {
			if strings.Contains(strings.ToUpper(s.Code), search) ||
				strings.Contains(strings.ToUpper(s.Label), search) {
				filtered = append(filtered, s)
			}
		}
		stations = filtered
	}

	if !output.IsInteractive() {
		output.JSON(map[string]any{"stations": stations, "count": len(stations), "source": source.SLF})
		return nil
	}

	output.Section(fmt.Sprintf("IMIS Stations (%d)", len(stations)))
	headers := []string{"CODE", "NAME", "ELEVATION", "CANTON", "TYPE"}
	var rows [][]string
	for _, s := range stations {
		rows = append(rows, []string{
			s.Code,
			s.Label,
			fmt.Sprintf("%.0fm", s.Elevation),
			s.CantonCode,
			s.Type,
		})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.SLF)
	return nil
}

func listStudyPlotStations(client *api.Client) error {
	stations, err := client.GetStudyPlotStations()
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if stationsSearch != "" {
		search := strings.ToUpper(stationsSearch)
		var filtered []api.StudyPlotStation
		for _, s := range stations {
			if strings.Contains(strings.ToUpper(s.Code), search) ||
				strings.Contains(strings.ToUpper(s.Label), search) {
				filtered = append(filtered, s)
			}
		}
		stations = filtered
	}

	if !output.IsInteractive() {
		output.JSON(map[string]any{"stations": stations, "count": len(stations), "source": source.SLF})
		return nil
	}

	output.Section(fmt.Sprintf("Study Plot Stations (%d)", len(stations)))
	headers := []string{"CODE", "NAME", "ELEVATION", "CANTON"}
	var rows [][]string
	for _, s := range stations {
		rows = append(rows, []string{
			s.Code,
			s.Label,
			fmt.Sprintf("%.0fm", s.Elevation),
			s.CantonCode,
		})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.SLF)
	return nil
}
```

- [ ] **Step 6: Write measurements command**

Create `internal/whiterisk/cmd/measurements.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/matthias/swisscli/internal/whiterisk/api"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	measurementType   string
	measurementPeriod int
)

func init() {
	rootCmd.AddCommand(measurementsCmd)
	measurementsCmd.Flags().StringVar(&measurementType, "type", "imis", "Station type: imis or study-plot")
	measurementsCmd.Flags().IntVar(&measurementPeriod, "period", 1, "Period in days: 1, 3, or 7")
}

var measurementsCmd = &cobra.Command{
	Use:   "measurements [station]",
	Short: "Snow and weather measurements",
	Long:  "Show measurements from IMIS or study plot stations. Specify station code (e.g. DAV2).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		station := strings.ToUpper(args[0])
		client := api.NewClient(Lang)

		if measurementType == "study-plot" {
			return showStudyPlotMeasurements(client, station)
		}
		return showIMISMeasurements(client, station)
	},
}

func showIMISMeasurements(client *api.Client, station string) error {
	measurements, err := client.GetIMISMeasurementsByStation(station, measurementPeriod)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if !output.IsInteractive() {
		output.JSON(map[string]any{"station": station, "measurements": measurements, "source": source.SLF})
		return nil
	}

	output.Section(fmt.Sprintf("IMIS Station: %s", station))
	headers := []string{"TIME", "TEMP", "HUMIDITY", "SNOW", "WIND", "GUSTS", "DIR"}
	var rows [][]string
	for _, m := range measurements {
		rows = append(rows, []string{
			m.MeasureDate,
			fmtFloat(m.TA30MinMean, "°C"),
			fmtFloat(m.RH30MinMean, "%"),
			fmtFloat(m.HS, " cm"),
			fmtFloat(m.VW30MinMean, " m/s"),
			fmtFloat(m.VW30MinMax, " m/s"),
			fmtFloat(m.DW30MinMean, "°"),
		})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.SLF)
	return nil
}

func showStudyPlotMeasurements(client *api.Client, station string) error {
	measurements, err := client.GetStudyPlotMeasurementsByStation(station, measurementPeriod)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if !output.IsInteractive() {
		output.JSON(map[string]any{"station": station, "measurements": measurements, "source": source.SLF})
		return nil
	}

	output.Section(fmt.Sprintf("Study Plot Station: %s", station))
	headers := []string{"TIME", "SNOW HEIGHT", "NEW SNOW 24h", "WATER EQ"}
	var rows [][]string
	for _, m := range measurements {
		rows = append(rows, []string{
			m.MeasureDate,
			fmtFloat(m.HS, " cm"),
			fmtFloat(m.HN1D, " cm"),
			fmtFloat(m.HNW1D, " mm"),
		})
	}
	output.Table(headers, rows)
	fmt.Printf("\n%s\n", source.SLF)
	return nil
}

func fmtFloat(v *float64, unit string) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.1f%s", *v, unit)
}
```

- [ ] **Step 7: Verify and test**

```bash
go run ./cmd/whiterisk stations
go run ./cmd/whiterisk stations --type study-plot
go run ./cmd/whiterisk measurements DAV2
go run ./cmd/whiterisk measurements DAV2 --period 3
go run ./cmd/whiterisk measurements DAV2 --json
```

- [ ] **Step 8: Commit**

```bash
git add internal/whiterisk/
git commit -m "feat: add whiterisk stations and measurements commands"
```

---

### Task 19: WhiteRisk snow maps command

**Files:**
- Create: `internal/whiterisk/api/snow.go`
- Create: `internal/whiterisk/cmd/snow.go`

- [ ] **Step 1: Write snow API**

Create `internal/whiterisk/api/snow.go`:

```go
package api

type SnowMapType string

const (
	SnowMapNew     SnowMapType = "new"
	SnowMapDepth   SnowMapType = "depth"
	SnowMapCompare SnowMapType = "compare"
)

var snowBrowserURLs = map[SnowMapType]string{
	SnowMapNew:     "https://whiterisk.ch/de/conditions/snow-maps/new_snow",
	SnowMapDepth:   "https://whiterisk.ch/de/conditions/snow-maps/snow_depth",
	SnowMapCompare: "https://whiterisk.ch/de/conditions/snow-maps/comparative_snow_depth",
}

var snowTeaserURLs = map[SnowMapType]string{
	SnowMapNew:     "https://whiterisk.ch/snowmap-teaser/new-snow.png",
	SnowMapDepth:   "https://whiterisk.ch/snowmap-teaser/snow-depth.png",
	SnowMapCompare: "https://whiterisk.ch/snowmap-teaser/comparative-snow-depth.png",
}

func GetSnowBrowserURL(t SnowMapType) string {
	return snowBrowserURLs[t]
}

func GetSnowTeaserURL(t SnowMapType) string {
	return snowTeaserURLs[t]
}
```

- [ ] **Step 2: Write snow command**

Create `internal/whiterisk/cmd/snow.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matthias/swisscli/internal/whiterisk/api"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

var (
	snowSave  string
	snowASCII bool
	snowWidth int
)

func init() {
	rootCmd.AddCommand(snowCmd)
	snowCmd.AddCommand(snowNewCmd)
	snowCmd.AddCommand(snowDepthCmd)
	snowCmd.AddCommand(snowCompareCmd)
	for _, c := range []*cobra.Command{snowNewCmd, snowDepthCmd, snowCompareCmd} {
		c.Flags().StringVar(&snowSave, "save", "", "Save image to file")
		c.Flags().BoolVar(&snowASCII, "ascii", false, "ASCII art mode")
		c.Flags().IntVar(&snowWidth, "width", 80, "ASCII art width")
	}
}

var snowCmd = &cobra.Command{
	Use:   "snow",
	Short: "Snow maps and data",
}

var snowNewCmd = &cobra.Command{
	Use:   "new",
	Short: "New snow map",
	RunE:  snowMapRunner(api.SnowMapNew),
}

var snowDepthCmd = &cobra.Command{
	Use:   "depth",
	Short: "Snow depth map",
	RunE:  snowMapRunner(api.SnowMapDepth),
}

var snowCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Comparative snow depth map",
	RunE:  snowMapRunner(api.SnowMapCompare),
}

func snowMapRunner(mapType api.SnowMapType) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		browserURL := api.GetSnowBrowserURL(mapType)
		imageURL := api.GetSnowTeaserURL(mapType)

		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"type":        string(mapType),
				"browser_url": browserURL,
				"image_url":   imageURL,
				"source":      source.SLF,
			})
			return nil
		}

		if snowASCII {
			fmt.Printf("Fetching %s snow map...\n", mapType)
			if err := output.ASCIIMap(imageURL, snowWidth); err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			fmt.Printf("\n%s\n", source.SLF)
			return nil
		}

		if snowSave != "" {
			path := snowSave
			if filepath.Ext(path) == "" {
				path += ".png"
			}
			fmt.Printf("Saving %s snow map to %s...\n", mapType, path)
			if err := output.SaveImage(imageURL, path); err != nil {
				output.Error(err.Error())
				os.Exit(1)
			}
			fmt.Printf("Saved to %s\n", path)
			fmt.Printf("\n%s\n", source.SLF)
			return nil
		}

		fmt.Printf("Opening %s snow map in browser...\n", mapType)
		if err := output.OpenBrowser(browserURL); err != nil {
			fmt.Printf("Visit: %s\n", browserURL)
		}
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	}
}
```

- [ ] **Step 3: Verify and test**

```bash
go run ./cmd/whiterisk snow new
go run ./cmd/whiterisk snow depth --ascii
go run ./cmd/whiterisk snow compare --json
```

- [ ] **Step 4: Commit**

```bash
git add internal/whiterisk/
git commit -m "feat: add whiterisk snow maps command with browser, save, ASCII modes"
```

---

### Task 20: WhiteRisk avalanches + profiles + favorites commands

**Files:**
- Create: `internal/whiterisk/cmd/avalanches.go`
- Create: `internal/whiterisk/cmd/profiles.go`
- Create: `internal/whiterisk/cmd/favorites.go`

- [ ] **Step 1: Write avalanches command**

Create `internal/whiterisk/cmd/avalanches.go`:

```go
package cmd

import (
	"fmt"

	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

const avalanchesURL = "https://whiterisk.ch/de/conditions/current-avalanches"

func init() {
	rootCmd.AddCommand(avalanchesCmd)
}

var avalanchesCmd = &cobra.Command{
	Use:   "avalanches",
	Short: "Current reported avalanches",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"url":    avalanchesURL,
				"source": source.SLF,
			})
			return nil
		}

		fmt.Println("Opening current avalanches in browser...")
		if err := output.OpenBrowser(avalanchesURL); err != nil {
			fmt.Printf("Visit: %s\n", avalanchesURL)
		}
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	},
}
```

- [ ] **Step 2: Write profiles command**

Create `internal/whiterisk/cmd/profiles.go`:

```go
package cmd

import (
	"fmt"

	"github.com/matthias/swisscli/pkg/output"
	"github.com/matthias/swisscli/pkg/source"
	"github.com/spf13/cobra"
)

const profilesURL = "https://whiterisk.ch/de/conditions/snow-profiles"

func init() {
	rootCmd.AddCommand(profilesCmd)
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Snow profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !output.IsInteractive() {
			output.JSON(map[string]string{
				"url":    profilesURL,
				"source": source.SLF,
			})
			return nil
		}

		fmt.Println("Opening snow profiles in browser...")
		if err := output.OpenBrowser(profilesURL); err != nil {
			fmt.Printf("Visit: %s\n", profilesURL)
		}
		fmt.Printf("\n%s\n", source.SLF)
		return nil
	},
}
```

- [ ] **Step 3: Write favorites command**

Create `internal/whiterisk/cmd/favorites.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/matthias/swisscli/pkg/config"
	"github.com/matthias/swisscli/pkg/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(favoritesCmd)
	favoritesCmd.AddCommand(favoritesAddCmd)
	favoritesCmd.AddCommand(favoritesRemoveCmd)
	favoritesCmd.AddCommand(favoritesListCmd)
}

var favoritesCmd = &cobra.Command{
	Use:   "favorites",
	Short: "Manage saved locations and stations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return favoritesListCmd.RunE(cmd, args)
	},
}

var favoritesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("whiterisk")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		if !output.IsInteractive() {
			output.JSON(cfg.Favorites)
			return nil
		}

		if len(cfg.Favorites) == 0 {
			fmt.Println("No favorites saved. Use `whiterisk favorites add <name> <station|region>` to add one.")
			return nil
		}

		headers := []string{"NAME", "STATION", "REGION"}
		var rows [][]string
		for _, f := range cfg.Favorites {
			rows = append(rows, []string{f.Name, f.Station, f.Region})
		}
		output.Table(headers, rows)
		return nil
	},
}

var favoritesAddCmd = &cobra.Command{
	Use:   "add <name> <station-or-region>",
	Short: "Add a favorite station or region",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		value := args[1]

		cfg, err := config.LoadDefault("whiterisk")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		fav := config.Favorite{Name: name}
		// If it looks like a station code (letters+digits, short), treat as station
		// Otherwise treat as region
		if len(value) <= 6 {
			fav.Station = value
		} else {
			fav.Region = value
		}

		cfg.AddFavorite(fav)
		if err := cfg.Save(); err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}
		fmt.Printf("Added %q to favorites.\n", name)
		return nil
	},
}

var favoritesRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a favorite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault("whiterisk")
		if err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}

		cfg.RemoveFavorite(args[0])
		if err := cfg.Save(); err != nil {
			output.Error(err.Error())
			os.Exit(1)
		}
		fmt.Printf("Removed %q from favorites.\n", args[0])
		return nil
	},
}
```

- [ ] **Step 4: Verify all whiterisk commands build**

```bash
go build ./cmd/whiterisk
./whiterisk --help
```

Expected: all commands listed.

- [ ] **Step 5: Commit**

```bash
git add internal/whiterisk/cmd/
git commit -m "feat: add whiterisk avalanches, profiles, and favorites commands"
```

---

## Phase 4: Final Integration

### Task 21: Run all tests + build both binaries

- [ ] **Step 1: Run all tests**

```bash
go test ./...
```

Expected: all tests PASS.

- [ ] **Step 2: Build both binaries**

```bash
make build
```

Expected: `meteoswiss` and `whiterisk` binaries created.

- [ ] **Step 3: Verify both CLIs end-to-end**

```bash
./meteoswiss --help
./meteoswiss forecast 8001
./meteoswiss current SMA
./meteoswiss stations --search TAE
./meteoswiss radar rain --json

./whiterisk --help
./whiterisk bulletin
./whiterisk stations --search DAV
./whiterisk measurements DAV2
./whiterisk snow new --json
```

- [ ] **Step 4: Add .gitignore**

Create `.gitignore`:

```
meteoswiss
whiterisk
dist/
*.exe
.DS_Store
```

- [ ] **Step 5: Final commit**

```bash
git add .gitignore
git commit -m "chore: add .gitignore and verify full build"
```

---

## Summary

| Phase | Tasks | What it builds |
|-------|-------|---------------|
| 1: Scaffolding | 1-6 | go.mod, Makefile, shared packages (source, output, config, geo) |
| 2: MeteoSwiss | 7-14 | Root cmd, API client, forecast, current, radar, stations, favorites, hazards, wind, pollen, precipitation, clouds |
| 3: WhiteRisk | 15-20 | Root cmd, API client, bulletin, stations, measurements, snow maps, avalanches, profiles, favorites |
| 4: Integration | 21 | Full test + build verification |
