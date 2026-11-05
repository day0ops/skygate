package rest

import "net/http"

type metaResponse struct {
	Version string `json:"version"`
	Color   string `json:"color"`
}

// RegisterMeta adds the /v1/meta (canary identity) and /v1/whoami (identity
// echo) endpoints to mux. whoami does not interpret or enforce anything -
// it only reflects whatever headers arrived, which is the point: it shows
// exactly what kgateway forwarded.
func RegisterMeta(mux *http.ServeMux, version, color string) {
	mux.HandleFunc("GET /v1/meta", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, metaResponse{Version: version, Color: color})
	})

	mux.HandleFunc("GET /v1/whoami", func(w http.ResponseWriter, r *http.Request) {
		// Keep all values per header, not just the first - a header sent
		// multiple times (e.g. a hop-by-hop forwarded-for chain) must be
		// shown in full to make good on "exactly what kgateway forwarded."
		headers := make(map[string][]string, len(r.Header))
		for k, v := range r.Header {
			headers[k] = v
		}
		writeJSON(w, http.StatusOK, map[string]any{"headers": headers})
	})
}
