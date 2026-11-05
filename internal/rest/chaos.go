package rest

import (
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// ChaosState holds the currently configured fault-injection parameters.
// Use NewChaosState to construct one - the zero value has an unset
// errorRate atomic.Value, which panics on Load.
type ChaosState struct {
	delay        atomic.Int64 // nanoseconds
	errorRate    atomic.Value // float64
	errorStatus  atomic.Int64
	panicEnabled bool
}

// NewChaosState returns a ChaosState with no latency/errors injected by
// default. The panic endpoint is gated by panicEnabled.
func NewChaosState(panicEnabled bool) *ChaosState {
	cs := &ChaosState{panicEnabled: panicEnabled}
	cs.errorRate.Store(0.0)
	cs.errorStatus.Store(int64(http.StatusInternalServerError))
	return cs
}

var chaosExemptPrefixes = []string{"/healthz", "/readyz", "/metrics", "/v1/chaos"}

func chaosExempt(path string) bool {
	for _, p := range chaosExemptPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// Middleware injects configured latency/errors into every non-exempt
// request, so resilience/observability demos can be triggered on demand.
func (cs *ChaosState) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if chaosExempt(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		if d := time.Duration(cs.delay.Load()); d > 0 {
			time.Sleep(d)
		}

		if rate, _ := cs.errorRate.Load().(float64); rate > 0 && rand.Float64() < rate {
			writeError(w, int(cs.errorStatus.Load()), "chaos: injected error")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RegisterChaos adds the chaos-control endpoints to mux.
func RegisterChaos(mux *http.ServeMux, cs *ChaosState) {
	mux.HandleFunc("POST /v1/chaos/delay", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			DurationMs int `json:"durationMs"`
		}
		if err := decodeJSON(r, &req); err != nil || req.DurationMs < 0 {
			writeError(w, http.StatusBadRequest, "durationMs must be a non-negative integer")
			return
		}
		cs.delay.Store(int64(time.Duration(req.DurationMs) * time.Millisecond))
		writeJSON(w, http.StatusOK, map[string]int{"durationMs": req.DurationMs})
	})

	mux.HandleFunc("POST /v1/chaos/error-rate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Rate   float64 `json:"rate"`
			Status int     `json:"status"`
		}
		if err := decodeJSON(r, &req); err != nil || req.Rate < 0 || req.Rate > 1 {
			writeError(w, http.StatusBadRequest, "rate must be between 0 and 1")
			return
		}
		status := req.Status
		if status == 0 {
			status = http.StatusInternalServerError
		}
		cs.errorRate.Store(req.Rate)
		cs.errorStatus.Store(int64(status))
		writeJSON(w, http.StatusOK, map[string]any{"rate": req.Rate, "status": status})
	})

	mux.HandleFunc("GET /v1/chaos/panic", func(w http.ResponseWriter, r *http.Request) {
		if !cs.panicEnabled {
			writeError(w, http.StatusForbidden, "panic endpoint disabled, set CHAOS_PANIC_ENABLED=true")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "crashing"})
		go func() {
			time.Sleep(100 * time.Millisecond) // let the response flush first
			os.Exit(1)
		}()
	})
}
