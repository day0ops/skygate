package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "skygate_http_requests_total",
			Help: "Total HTTP requests, labeled by route, method, status and version.",
		},
		[]string{"route", "method", "status", "version"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "skygate_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds, labeled by route and version.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route", "version"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal, requestDuration)
}

// Handler serves the Prometheus exposition format.
func Handler() http.Handler {
	return promhttp.Handler()
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Instrument wraps next, recording RED metrics for every request. The route
// label comes from r.Pattern (the matched http.ServeMux pattern, e.g.
// "GET /v1/flights/{id}"), stripped of its method prefix, so cardinality
// stays bounded regardless of how many distinct flight IDs are requested.
//
// A handler panic is recorded as a 500 and re-raised for net/http's own
// recovery to handle, so a crashing handler still shows up in RED metrics
// instead of vanishing from them - exactly the failure these metrics exist
// to surface.
func Instrument(version string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		// r.Pattern is only populated once the mux has matched and started
		// dispatching to next, so it must be read inside this deferred
		// closure (after next.ServeHTTP runs, whether it returns normally
		// or panics) - reading it before the call always sees it empty.
		record := func(status int) {
			route := r.Pattern
			if idx := strings.IndexByte(route, ' '); idx >= 0 {
				route = route[idx+1:]
			}
			if route == "" {
				route = r.URL.Path
			}
			requestsTotal.WithLabelValues(route, r.Method, strconv.Itoa(status), version).Inc()
			requestDuration.WithLabelValues(route, version).Observe(time.Since(start).Seconds())
		}

		defer func() {
			if p := recover(); p != nil {
				record(http.StatusInternalServerError)
				panic(p)
			}
			record(rec.status)
		}()

		next.ServeHTTP(rec, r)
	})
}
