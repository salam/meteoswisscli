//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binPath string

func TestMain(m *testing.M) {
	// Build the binary
	tmpDir, err := os.MkdirTemp("", "meteoswiss-e2e-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	binPath = filepath.Join(tmpDir, "meteoswiss")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/meteoswiss")
	build.Dir = filepath.Join("..")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("failed to build: " + err.Error())
	}

	os.Exit(m.Run())
}

func skipIfNoNetwork(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("set INTEGRATION_TEST=1 to run e2e tests with network")
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command(binPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	if !strings.Contains(string(out), "meteoswiss") {
		t.Errorf("version output doesn't contain binary name: %s", out)
	}
}

func TestHelpFlag(t *testing.T) {
	cmd := exec.Command(binPath, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}
	output := string(out)
	for _, sub := range []string{"forecast", "current", "wind", "pollen", "bulletin", "stations", "radar"} {
		if !strings.Contains(output, sub) {
			t.Errorf("help should mention %q subcommand", sub)
		}
	}
}

func TestForecastHelp(t *testing.T) {
	cmd := exec.Command(binPath, "forecast", "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("forecast --help failed: %v", err)
	}
	if !strings.Contains(string(out), "location") {
		t.Error("forecast help should mention location argument")
	}
}

func TestForecastMissingArg(t *testing.T) {
	cmd := exec.Command(binPath, "forecast")
	err := cmd.Run()
	if err == nil {
		t.Error("forecast without args should exit with error")
	}
}

func TestUnknownCommand(t *testing.T) {
	cmd := exec.Command(binPath, "nonexistent")
	err := cmd.Run()
	if err == nil {
		t.Error("unknown command should exit with error")
	}
}

func TestForecastJSON_E2E(t *testing.T) {
	skipIfNoNetwork(t)

	cmd := exec.Command(binPath, "forecast", "8001", "--json")
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("forecast failed (network?): %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["source"]; !ok {
		t.Error("missing source field")
	}
	if _, ok := result["currentWeather"]; !ok {
		t.Error("missing currentWeather field")
	}
}

func TestCurrentJSON_E2E(t *testing.T) {
	skipIfNoNetwork(t)

	cmd := exec.Command(binPath, "current", "--json")
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("current failed (network?): %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["measurements"]; !ok {
		t.Error("missing measurements field")
	}
}

func TestWindBrowserJSON_E2E(t *testing.T) {
	cmd := exec.Command(binPath, "wind", "--browser", "--json")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("wind --browser --json failed: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(out, &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if result["url"] == "" {
		t.Error("missing url field")
	}
}

func TestRadarJSON_E2E(t *testing.T) {
	cmd := exec.Command(binPath, "radar", "--json")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("radar --json failed: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(out, &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if result["url"] == "" {
		t.Error("missing url field")
	}
}

func TestStationsNearJSON_E2E(t *testing.T) {
	cmd := exec.Command(binPath, "stations", "--near", "Zürich", "--json")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("stations --near --json failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["stations"]; !ok {
		t.Error("missing stations field")
	}
}

func TestBulletinJSON_E2E(t *testing.T) {
	skipIfNoNetwork(t)

	cmd := exec.Command(binPath, "bulletin", "--json")
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("bulletin failed (network?): %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if _, ok := result["text"]; !ok {
		t.Error("missing text field")
	}
}

func TestPollenJSON_E2E(t *testing.T) {
	skipIfNoNetwork(t)

	cmd := exec.Command(binPath, "pollen", "PZH", "--json")
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("pollen failed (network?): %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
		return
	}
	if result["station"] != "PZH" {
		t.Errorf("station = %v, want PZH", result["station"])
	}
}

func TestLangFlag_E2E(t *testing.T) {
	skipIfNoNetwork(t)

	cmd := exec.Command(binPath, "forecast", "8001", "--json", "--lang", "fr")
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("forecast with --lang failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}
}

func TestExitCodeSuccess(t *testing.T) {
	cmd := exec.Command(binPath, "--version")
	if err := cmd.Run(); err != nil {
		t.Errorf("--version should exit with code 0, got: %v", err)
	}
}

func TestExitCodeError(t *testing.T) {
	cmd := exec.Command(binPath, "forecast")
	if err := cmd.Run(); err == nil {
		t.Error("forecast without args should exit with non-zero code")
	}
}
