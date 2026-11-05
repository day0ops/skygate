package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetaReportsVersionAndColor(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMeta(mux, "v1.2.3", "green")

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	if !contains(rec.Body.String(), "v1.2.3") || !contains(rec.Body.String(), "green") {
		t.Fatalf("got body %s, want it to contain version and color", rec.Body.String())
	}
}

func TestWhoamiEchoesHeaders(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMeta(mux, "v1.2.3", "green")

	req := httptest.NewRequest(http.MethodGet, "/v1/whoami", nil)
	req.Header.Set("X-Forwarded-User", "ada@example.com")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	if !contains(rec.Body.String(), "ada@example.com") {
		t.Fatalf("got body %s, want it to echo the forwarded-user header", rec.Body.String())
	}
}
