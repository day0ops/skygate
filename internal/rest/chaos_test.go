package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChaosDelayInjectsLatency(t *testing.T) {
	cs := NewChaosState(false)
	mux := http.NewServeMux()
	RegisterChaos(mux, cs)
	inner := http.NewServeMux()
	inner.HandleFunc("GET /probe", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := cs.Middleware(inner)

	body := strings.NewReader(`{"durationMs":50}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/chaos/delay", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200, body=%s", rec.Code, rec.Body.String())
	}

	start := time.Now()
	req2 := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Fatalf("expected at least 50ms delay, took %v", elapsed)
	}
}

func TestChaosErrorRateInjectsErrors(t *testing.T) {
	cs := NewChaosState(false)
	mux := http.NewServeMux()
	RegisterChaos(mux, cs)
	inner := http.NewServeMux()
	inner.HandleFunc("GET /probe", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := cs.Middleware(inner)

	body := strings.NewReader(`{"rate":1,"status":503}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/chaos/error-rate", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200, body=%s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("got status %d, want 503 (rate=1 must always inject)", rec2.Code)
	}
}

func TestChaosPanicDisabledByDefault(t *testing.T) {
	cs := NewChaosState(false)
	mux := http.NewServeMux()
	RegisterChaos(mux, cs)

	req := httptest.NewRequest(http.MethodGet, "/v1/chaos/panic", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want 403 when panicEnabled=false", rec.Code)
	}
}

func TestChaosMiddlewareExemptsHealthAndChaosRoutes(t *testing.T) {
	cs := NewChaosState(false)
	body := strings.NewReader(`{"rate":1,"status":503}`)
	setReq := httptest.NewRequest(http.MethodPost, "/v1/chaos/error-rate", body)
	setRec := httptest.NewRecorder()
	chaosMux := http.NewServeMux()
	RegisterChaos(chaosMux, cs)
	chaosMux.ServeHTTP(setRec, setReq)

	inner := http.NewServeMux()
	RegisterHealth(inner)
	handler := cs.Middleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 - health checks must be exempt from chaos", rec.Code)
	}
}

func TestChaosMiddlewareWrapsRealRouterWiring(t *testing.T) {
	cs := NewChaosState(false)
	inner := http.NewServeMux()
	RegisterHealth(inner)
	RegisterChaos(inner, cs)
	wrapped := http.NewServeMux()
	wrapped.Handle("/", cs.Middleware(inner))

	// Set rate=1 through the WRAPPED handler, the same path a real client uses.
	body := strings.NewReader(`{"rate":1,"status":503}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/chaos/error-rate", body)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 setting error-rate, body=%s", rec.Code, rec.Body.String())
	}

	// /healthz must still be exempt even through the full wrapped stack.
	req2 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 - health must stay exempt through the real wiring", rec2.Code)
	}

	// The chaos-control endpoint itself must also stay exempt (so you can always turn chaos back off).
	body2 := strings.NewReader(`{"rate":0}`)
	req3 := httptest.NewRequest(http.MethodPost, "/v1/chaos/error-rate", body2)
	rec3 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 resetting error-rate through the wrapped stack", rec3.Code)
	}
}
