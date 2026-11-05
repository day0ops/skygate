package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInstrumentRecordsRequest(t *testing.T) {
	inner := http.NewServeMux()
	inner.HandleFunc("GET /v1/airports", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Instrument("v1-test", inner)

	req := httptest.NewRequest(http.MethodGet, "/v1/airports", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	Handler().ServeHTTP(metricsRec, metricsReq)

	body := metricsRec.Body.String()
	want := `skygate_http_requests_total{method="GET",route="/v1/airports",status="200",version="v1-test"}`
	if !strings.Contains(body, want) {
		t.Fatalf("expected metrics to contain %q, got:\n%s", want, body)
	}
}

func TestInstrumentRecordsPanicAs500(t *testing.T) {
	inner := http.NewServeMux()
	inner.HandleFunc("GET /v1/chaos/panic-test", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	handler := Instrument("v1-panic-test", inner)

	req := httptest.NewRequest(http.MethodGet, "/v1/chaos/panic-test", nil)
	rec := httptest.NewRecorder()

	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Fatal("expected Instrument to re-panic after recording metrics")
			}
		}()
		handler.ServeHTTP(rec, req)
	}()

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	Handler().ServeHTTP(metricsRec, metricsReq)

	body := metricsRec.Body.String()
	want := `skygate_http_requests_total{method="GET",route="/v1/chaos/panic-test",status="500",version="v1-panic-test"}`
	if !strings.Contains(body, want) {
		t.Fatalf("expected a panic to be recorded as a 500, got:\n%s", body)
	}
}

func TestInstrumentUsesPatternNotRawPathForCardinality(t *testing.T) {
	inner := http.NewServeMux()
	inner.HandleFunc("GET /v1/flights-cardinality-test/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Instrument("v1-cardinality-test", inner)

	for _, id := range []string{"flt_1", "flt_2"} {
		req := httptest.NewRequest(http.MethodGet, "/v1/flights-cardinality-test/"+id, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	Handler().ServeHTTP(metricsRec, metricsReq)

	body := metricsRec.Body.String()
	want := `skygate_http_requests_total{method="GET",route="/v1/flights-cardinality-test/{id}",status="200",version="v1-cardinality-test"} 2`
	if !strings.Contains(body, want) {
		t.Fatalf("expected both requests counted under the pattern route (not per-ID raw paths), got:\n%s", body)
	}
	if strings.Contains(body, `route="/v1/flights-cardinality-test/flt_1"`) || strings.Contains(body, `route="/v1/flights-cardinality-test/flt_2"`) {
		t.Fatalf("expected no per-ID route labels (cardinality leak), got:\n%s", body)
	}
}
