package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/salam/swissmeteocli/pkg/output"
)

func init() {
	// Ensure commands are not colored for consistent test output
	output.NoColor = true
}

func skipIfNoIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("set INTEGRATION_TEST=1 to run integration tests")
	}
}

// executeCommand captures both cobra output and direct stdout writes.
func executeCommand(args ...string) (string, error) {
	// Capture stdout since commands write directly via fmt.Printf and output.JSON
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cobraBuf := new(bytes.Buffer)
	rootCmd.SetOut(cobraBuf)
	rootCmd.SetErr(cobraBuf)
	rootCmd.SetArgs(args)

	// Reset flags to defaults before each test
	output.ForceJSON = false
	weekFlag = false

	// Read pipe in background to prevent deadlock on large output
	var stdoutBuf bytes.Buffer
	done := make(chan struct{})
	go func() {
		stdoutBuf.ReadFrom(r)
		close(done)
	}()

	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout
	<-done

	// Combine cobra output and stdout output
	combined := cobraBuf.String() + stdoutBuf.String()
	return combined, err
}

// --- Forecast tests ---

func TestForecastJSON_ByPLZ(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("forecast", "8001", "--json")
	if err != nil {
		t.Skipf("API call failed (network required): %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON output: %v\noutput: %s", err, out)
		return
	}
	if _, ok := result["source"]; !ok {
		t.Error("JSON output missing 'source' field")
	}
	if _, ok := result["currentWeather"]; !ok {
		t.Error("JSON output missing 'currentWeather' field")
	}
	if _, ok := result["forecast"]; !ok {
		t.Error("JSON output missing 'forecast' field")
	}
}

func TestForecastJSON_ByName(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("forecast", "Zürich", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["location"]; !ok {
		t.Error("JSON output missing 'location' field for named location")
	}
}

func TestForecastTable(t *testing.T) {
	skipIfNoIntegration(t)

	// When captured via pipe, IsInteractive() is false, so output is JSON.
	// Verify JSON output contains forecast data with temperature values.
	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("forecast", "8001", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v\noutput: %s", err, out)
		return
	}
	forecasts, ok := result["forecast"].([]any)
	if !ok || len(forecasts) == 0 {
		t.Error("forecast should contain at least one day")
	}
}

// --- Current tests ---

func TestCurrentJSON_NoArgs(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("current", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v\noutput: %s", err, out)
		return
	}
	if _, ok := result["measurements"]; !ok {
		t.Error("JSON output missing 'measurements' field")
	}
	if _, ok := result["source"]; !ok {
		t.Error("JSON output missing 'source' field")
	}
}

func TestCurrentJSON_WithStation(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("current", "SMA", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	measurements, ok := result["measurements"].([]any)
	if !ok {
		t.Error("measurements should be an array")
		return
	}
	if len(measurements) == 0 {
		t.Error("should have at least one measurement for SMA")
	}
}

// --- Wind tests ---

func TestWindJSON(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("wind", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["measurements"]; !ok {
		t.Error("JSON output missing 'measurements' field")
	}
}

func TestWindJSON_BrowserMode(t *testing.T) {
	// --browser --json should output URL as JSON without opening browser
	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("wind", "--browser", "--json")
	if err != nil {
		t.Fatalf("wind --browser --json failed: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v\noutput: %s", err, out)
		return
	}
	if result["url"] == "" {
		t.Error("JSON output missing 'url' field")
	}
}

// --- Pollen tests ---

func TestPollenJSON(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("pollen", "PZH", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if result["station"] != "PZH" {
		t.Errorf("station = %v, want PZH", result["station"])
	}
	if _, ok := result["measurements"]; !ok {
		t.Error("JSON output missing 'measurements' field")
	}
}

func TestPollenJSON_AllStations(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("pollen", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["stations"]; !ok {
		t.Error("JSON output missing 'stations' field")
	}
}

// --- Bulletin tests ---

func TestBulletinJSON(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("bulletin", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["text"]; !ok {
		t.Error("JSON output missing 'text' field")
	}
	if _, ok := result["region"]; !ok {
		t.Error("JSON output missing 'region' field")
	}
	if _, ok := result["source"]; !ok {
		t.Error("JSON output missing 'source' field")
	}
}

func TestBulletinJSON_SouthRegion(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("bulletin", "--region", "south", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if result["region"] != "south" {
		t.Errorf("region = %v, want south", result["region"])
	}
}

// --- Stations tests ---

func TestStationsJSON(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("stations", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["stations"]; !ok {
		t.Error("JSON output missing 'stations' field")
	}
	if _, ok := result["count"]; !ok {
		t.Error("JSON output missing 'count' field")
	}
}

func TestStationsJSON_Near(t *testing.T) {
	// --near uses embedded data, no network required
	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("stations", "--near", "Bern", "--json")
	if err != nil {
		t.Fatalf("stations --near failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v\noutput: %s", err, out)
		return
	}
	if _, ok := result["stations"]; !ok {
		t.Error("JSON output missing 'stations' field")
	}
	if _, ok := result["near"]; !ok {
		t.Error("JSON output missing 'near' field")
	}
}

// --- Radar tests ---

func TestRadarJSON_Default(t *testing.T) {
	// Default radar mode (browser URL) should work without network
	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("radar", "--json")
	if err != nil {
		t.Fatalf("radar --json failed: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v\noutput: %s", err, out)
		return
	}
	if result["url"] == "" {
		t.Error("JSON output missing 'url' field")
	}
	if result["type"] == "" {
		t.Error("JSON output missing 'type' field")
	}
}

func TestRadarListJSON(t *testing.T) {
	skipIfNoIntegration(t)

	output.ForceJSON = true
	defer func() { output.ForceJSON = false }()

	out, err := executeCommand("radar", "--list", "--json")
	if err != nil {
		t.Skipf("API call failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["frames"]; !ok {
		t.Error("JSON output missing 'frames' field")
	}
}

// --- Version test ---

func TestVersion(t *testing.T) {
	out, err := executeCommand("--version")
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	if !strings.Contains(out, "meteoswiss") {
		t.Errorf("version output should contain 'meteoswiss', got: %s", out)
	}
}

// --- Help test ---

func TestHelp(t *testing.T) {
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}
	if !strings.Contains(out, "forecast") {
		t.Error("help should mention forecast command")
	}
	if !strings.Contains(out, "current") {
		t.Error("help should mention current command")
	}
	if !strings.Contains(out, "radar") {
		t.Error("help should mention radar command")
	}
}

// --- Error cases ---
// Note: Commands that call os.Exit(1) on error cannot be tested in-process.
// Those are covered by e2e tests which run the binary as a subprocess.
