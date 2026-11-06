package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetaReportsVersionColorAndBuildVersion(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMeta(mux, "v1.2.3", "green", "v0.3.0")

	req := httptest.NewRequest(http.MethodGet, "/v1/meta", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !contains(body, "v1.2.3") || !contains(body, "green") {
		t.Fatalf("got body %s, want it to contain version and color", body)
	}
	// buildVersion is the actual software release (e.g. v0.3.0), distinct from
	// version (a canary demo label like v1/v2 that routing-canary overrides) -
	// the two must never collapse into the same field.
	if !contains(body, "v0.3.0") {
		t.Fatalf("got body %s, want it to contain buildVersion", body)
	}
}

func TestWhoamiEchoesHeaders(t *testing.T) {
	mux := http.NewServeMux()
	RegisterMeta(mux, "v1.2.3", "green", "v0.3.0")

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
