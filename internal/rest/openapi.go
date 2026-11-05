package rest

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openapiSpec []byte

// RegisterOpenAPI serves the static OpenAPI document describing the public
// REST surface - this is what the Developer Portal demo publishes.
func RegisterOpenAPI(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(openapiSpec)
	})
}
