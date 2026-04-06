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
