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
