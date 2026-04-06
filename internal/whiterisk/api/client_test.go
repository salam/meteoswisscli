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
	err := c.DoJSON("GET", server.URL+"/test", &result)
	if err != nil {
		t.Fatalf("DoJSON error = %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %q, want ok", result["status"])
	}
}
