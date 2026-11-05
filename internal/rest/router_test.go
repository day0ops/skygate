package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"skygate/internal/domain"
	"skygate/internal/store"
)

// TestNewMuxSurvivesAdditionalCatchAllRoute mirrors the exact composition
// main.go performs: call NewMux, register more routes (including a
// method-specific catch-all like "GET /", the way the bundled UI's static
// file server does) directly on the returned mux, then wrap the whole thing
// with ChaosState.Middleware exactly once, at the end.
//
// Wrapping NewMux's result BEFORE adding more routes - as a prior version
// of this package did internally - breaks this: once something registers
// "GET /" on an already-wrapped mux, that pattern is more specific than the
// wrapper's own bare "/" and silently swallows every GET request the real
// API handlers were supposed to serve, including this test's own health
// check, well before any browser or curl call would ever reveal it.
func TestNewMuxSurvivesAdditionalCatchAllRoute(t *testing.T) {
	st := store.New()
	st.PutAirport(domain.Airport{Code: "SFO", Name: "San Francisco Intl"})

	mux, cs := NewMux(Deps{Store: st, Version: "v1", Color: "blue"})

	mux.Handle("GET /ws/flights", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusSwitchingProtocols)
	}))
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("static file not found"))
	}))

	handler := cs.Middleware(mux)

	for _, path := range []string{"/healthz", "/v1/airports"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s: got status %d, want 200 (got body %q) - a catch-all route is shadowing the real API", path, rec.Code, rec.Body.String())
		}
	}

	// The catch-all must still work for a path no REST handler owns.
	req := httptest.NewRequest(http.MethodGet, "/some-static-asset.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /some-static-asset.js: got status %d, want 404 from the catch-all", rec.Code)
	}
}
