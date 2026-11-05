package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOpenAPISpecIsValidAndServed(t *testing.T) {
	var doc map[string]any
	if err := yaml.Unmarshal(openapiSpec, &doc); err != nil {
		t.Fatalf("embedded openapi.yaml does not parse: %v", err)
	}
	if _, ok := doc["openapi"]; !ok {
		t.Fatal("expected top-level \"openapi\" key")
	}
	if _, ok := doc["paths"]; !ok {
		t.Fatal("expected top-level \"paths\" key")
	}

	mux := http.NewServeMux()
	RegisterOpenAPI(mux)
	req := httptest.NewRequest(http.MethodGet, "/v1/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
}

// publicRESTRoutes lists every product-facing REST endpoint (path relative
// to the "/v1" server URL, HTTP method) that must appear in openapi.yaml -
// this is the surface the Developer Portal demo publishes. Intentionally
// excluded: /healthz, /readyz, /metrics (infra, not product API) and
// /v1/chaos/* (operator-only demo controls, not customer-facing).
//
// Update this list in the same change whenever a route is added to or
// removed from airports.go/flights.go/bookings.go/meta.go - that's what
// catches drift between the real API and the published spec (this test
// caught exactly that drift once already: POST /flights and PATCH
// /flights/{id}/status were implemented but never added to openapi.yaml).
var publicRESTRoutes = []struct {
	path   string
	method string
}{
	{"/airports", "get"},
	{"/airports/{code}", "get"},
	{"/flights", "get"},
	{"/flights", "post"},
	{"/flights/{id}", "get"},
	{"/flights/{id}/position", "get"},
	{"/flights/{id}/status", "patch"},
	{"/bookings", "post"},
	{"/bookings/{id}", "get"},
	{"/bookings/{id}", "delete"},
	{"/meta", "get"},
	{"/whoami", "get"},
}

func TestOpenAPIDocumentsEveryPublicRoute(t *testing.T) {
	var doc struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(openapiSpec, &doc); err != nil {
		t.Fatalf("embedded openapi.yaml does not parse: %v", err)
	}

	for _, route := range publicRESTRoutes {
		methods, ok := doc.Paths[route.path]
		if !ok {
			t.Errorf("openapi.yaml is missing path %q", route.path)
			continue
		}
		if _, ok := methods[route.method]; !ok {
			t.Errorf("openapi.yaml path %q is missing method %q", route.path, route.method)
		}
	}
}
