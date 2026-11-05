package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterHealth(t *testing.T) {
	mux := http.NewServeMux()
	RegisterHealth(mux)

	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: got status %d, want 200", path, rec.Code)
		}
		if rec.Body.String() != "ok" {
			t.Fatalf("%s: got body %q, want %q", path, rec.Body.String(), "ok")
		}
	}
}
