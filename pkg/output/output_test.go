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

func TestIsInteractive_Default(t *testing.T) {
	old := ForceJSON
	defer func() { ForceJSON = old }()

	ForceJSON = false
	// In test environment, stdout is typically not a terminal
	// So IsInteractive should return false (piped output)
	result := IsInteractive()
	// We don't assert the exact value since it depends on test runner,
	// but we verify it doesn't panic
	_ = result
}

func TestTable_NoColor(t *testing.T) {
	oldColor := NoColor
	defer func() { NoColor = oldColor }()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	NoColor = true
	Table([]string{"A", "B"}, [][]string{{"1", "2"}})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if strings.Contains(out, "\033[") {
		t.Error("NoColor=true should not produce ANSI escape codes")
	}
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Errorf("Table should contain headers, got: %s", out)
	}
}

func TestTable_MultipleRows(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	NoColor = true
	Table(
		[]string{"NAME", "VALUE", "UNIT"},
		[][]string{
			{"temp", "20", "°C"},
			{"wind", "15", "km/h"},
			{"rain", "0.5", "mm"},
		},
	)
	NoColor = false

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	for _, expected := range []string{"temp", "wind", "rain", "°C", "km/h"} {
		if !strings.Contains(out, expected) {
			t.Errorf("Table output should contain %q, got: %s", expected, out)
		}
	}
}

func TestJSON_NestedStructure(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	JSON(map[string]any{
		"source":  "test",
		"data":    []int{1, 2, 3},
		"nested":  map[string]string{"key": "val"},
	})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON output should be valid JSON: %v", err)
	}
	if result["source"] != "test" {
		t.Errorf("source = %v, want test", result["source"])
	}
	data, ok := result["data"].([]any)
	if !ok || len(data) != 3 {
		t.Errorf("data should be array of 3 elements, got %v", result["data"])
	}
}

func TestJSONError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	JSONError("test error")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSONError output should be valid JSON: %v", err)
	}
	if result["error"] != "test error" {
		t.Errorf("error = %q, want %q", result["error"], "test error")
	}
}

func TestSection(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Section("Test Title")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "Test Title") {
		t.Errorf("Section output should contain title, got: %s", out)
	}
	if !strings.Contains(out, "---") {
		t.Errorf("Section output should contain dashes, got: %s", out)
	}
}
