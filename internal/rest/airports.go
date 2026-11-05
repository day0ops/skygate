package rest

import (
	"net/http"

	"skygate/internal/store"
)

// RegisterAirports adds the read-only airport endpoints to mux.
func RegisterAirports(mux *http.ServeMux, st *store.Store) {
	mux.HandleFunc("GET /v1/airports", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, st.ListAirports())
	})

	mux.HandleFunc("GET /v1/airports/{code}", func(w http.ResponseWriter, r *http.Request) {
		a, ok := st.GetAirport(r.PathValue("code"))
		if !ok {
			writeError(w, http.StatusNotFound, "airport not found")
			return
		}
		writeJSON(w, http.StatusOK, a)
	})
}
