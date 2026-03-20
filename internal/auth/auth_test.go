package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if hash == "testpassword" {
		t.Error("hash should not equal plaintext")
	}
}

func TestHashPasswordDifferentEachTime(t *testing.T) {
	hash1, _ := HashPassword("same")
	hash2, _ := HashPassword("same")
	if hash1 == hash2 {
		t.Error("bcrypt should produce different hashes for same input")
	}
}

func TestRequireAuth_NoHeader(t *testing.T) {
	a := &Auth{}
	handler := a.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	a := &Auth{}
	handler := a.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
