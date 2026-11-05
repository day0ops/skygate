package rest

import (
	"net/http"

	"skygate/internal/metrics"
	"skygate/internal/store"
)

// Deps bundles everything the REST handlers need to be constructed.
type Deps struct {
	Store        *store.Store
	Version      string
	Color        string
	PanicEnabled bool
}

// NewMux builds the REST API's routes onto one http.ServeMux and returns it
// along with the ChaosState those routes were registered against.
//
// It deliberately does NOT wrap the mux in ChaosState.Middleware itself.
// http.ServeMux resolves a request against every pattern registered on it,
// regardless of how many separate calls registered them - so a caller that
// needs to add more routes later (the WebSocket feed, the bundled UI) must
// register them on this SAME mux before any wrapping happens. Wrapping here
// and then adding a method-specific pattern like "GET /" to the wrapped mux
// from the outside does not compose: the new pattern is more specific than
// the wrapper's own bare "/" registration and silently out-competes it for
// every GET request, hijacking the whole REST API behind a 404 from
// whatever that new route serves. Call Middleware on the final, fully
// assembled mux instead - see cmd/skygate/main.go.
func NewMux(d Deps) (*http.ServeMux, *ChaosState) {
	mux := http.NewServeMux()
	RegisterHealth(mux)
	RegisterAirports(mux, d.Store)
	RegisterFlights(mux, d.Store)
	RegisterBookings(mux, d.Store)
	RegisterMeta(mux, d.Version, d.Color)
	RegisterOpenAPI(mux)
	mux.Handle("GET /metrics", metrics.Handler())

	cs := NewChaosState(d.PanicEnabled)
	RegisterChaos(mux, cs)

	return mux, cs
}
