package retro

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJsonResponse(t *testing.T) {
	w := httptest.NewRecorder()
	jsonResponse(w, http.StatusCreated, map[string]string{"name": "test"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", w.Header().Get("Content-Type"))
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp["name"] != "test" {
		t.Errorf("expected 'test', got '%s'", resp["name"])
	}
}

func TestJsonError(t *testing.T) {
	w := httptest.NewRecorder()
	jsonError(w, fmt.Errorf("test error"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", w.Header().Get("Content-Type"))
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error != "test error" {
		t.Errorf("expected 'test error', got '%s'", resp.Error)
	}
}
