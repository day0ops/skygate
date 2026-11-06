package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"skygate/internal/domain"
	"skygate/internal/store"
)

func newFlightsTestMux(st *store.Store) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterFlights(mux, st)
	return mux
}

func TestListFlightsFiltersByStatus(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusScheduled, OriginCode: "SFO"})
	st.PutFlight(domain.Flight{ID: "flt_2", Status: domain.StatusAirborne, OriginCode: "JFK"})
	mux := newFlightsTestMux(st)

	req := httptest.NewRequest(http.MethodGet, "/v1/flights?status=airborne", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	if !contains(rec.Body.String(), "flt_2") || contains(rec.Body.String(), "flt_1") {
		t.Fatalf("got body %s, want only flt_2", rec.Body.String())
	}
}

func TestListFlightsFiltersByOrigin(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", OriginCode: "SFO", DestinationCode: "LAX"})
	st.PutFlight(domain.Flight{ID: "flt_2", OriginCode: "JFK", DestinationCode: "LAX"})
	mux := newFlightsTestMux(st)

	req := httptest.NewRequest(http.MethodGet, "/v1/flights?origin=SFO", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	if !contains(rec.Body.String(), "flt_1") || contains(rec.Body.String(), "flt_2") {
		t.Fatalf("got body %s, want only flt_1", rec.Body.String())
	}
}

func TestListFlightsFiltersByDestination(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", OriginCode: "SFO", DestinationCode: "LAX"})
	st.PutFlight(domain.Flight{ID: "flt_2", OriginCode: "SFO", DestinationCode: "JFK"})
	mux := newFlightsTestMux(st)

	req := httptest.NewRequest(http.MethodGet, "/v1/flights?destination=LAX", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	if !contains(rec.Body.String(), "flt_1") || contains(rec.Body.String(), "flt_2") {
		t.Fatalf("got body %s, want only flt_1", rec.Body.String())
	}
}

func TestGetFlightNotFound(t *testing.T) {
	mux := newFlightsTestMux(store.New())
	req := httptest.NewRequest(http.MethodGet, "/v1/flights/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestGetFlight(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", FlightNumber: "AA100"})
	mux := newFlightsTestMux(st)

	req := httptest.NewRequest(http.MethodGet, "/v1/flights/flt_1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	if !contains(rec.Body.String(), "flt_1") {
		t.Fatalf("got body %s, want to contain flt_1", rec.Body.String())
	}
}

func TestGetFlightPositionRequiresAirborne(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusScheduled})
	mux := newFlightsTestMux(st)

	req := httptest.NewRequest(http.MethodGet, "/v1/flights/flt_1/position", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("got status %d, want 409", rec.Code)
	}
}

func TestGetFlightPosition(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusAirborne, Position: &domain.Position{X: 10, Y: 20, Progress: 0.5}})
	mux := newFlightsTestMux(st)

	req := httptest.NewRequest(http.MethodGet, "/v1/flights/flt_1/position", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}
	if !contains(rec.Body.String(), "10") {
		t.Fatalf("got body %s, want to contain position data with 10", rec.Body.String())
	}
}

func TestCreateFlight(t *testing.T) {
	mux := newFlightsTestMux(store.New())
	body := strings.NewReader(`{"flightNumber":"SK999","originCode":"SFO","destinationCode":"JFK"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/flights", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, want 201, body=%s", rec.Code, rec.Body.String())
	}
	if !contains(rec.Body.String(), "SK999") {
		t.Fatalf("expected response to contain SK999, got %s", rec.Body.String())
	}
}

func TestCreateFlightRejectsMissingFields(t *testing.T) {
	mux := newFlightsTestMux(store.New())
	body := strings.NewReader(`{"flightNumber":""}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/flights", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestUpdateFlightStatus(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusScheduled})
	mux := newFlightsTestMux(st)

	body := strings.NewReader(`{"status":"cancelled"}`)
	req := httptest.NewRequest(http.MethodPatch, "/v1/flights/flt_1/status", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	f, _ := st.GetFlight("flt_1")
	if f.Status != domain.StatusCancelled {
		t.Fatalf("got status %q, want cancelled", f.Status)
	}
}

func TestUpdateFlightStatusPreservesPositionWhenHolding(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{
		ID:       "flt_1",
		Status:   domain.StatusAirborne,
		Position: &domain.Position{X: 10, Y: 20, Progress: 0.9},
	})
	mux := newFlightsTestMux(st)

	body := strings.NewReader(`{"status":"holding"}`)
	req := httptest.NewRequest(http.MethodPatch, "/v1/flights/flt_1/status", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	f, _ := st.GetFlight("flt_1")
	if f.Position == nil {
		t.Fatal("expected Position to be preserved when transitioning to holding")
	}
}

func TestUpdateFlightStatusRejectsUnknownStatus(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusScheduled})
	mux := newFlightsTestMux(st)

	body := strings.NewReader(`{"status":"orbiting"}`)
	req := httptest.NewRequest(http.MethodPatch, "/v1/flights/flt_1/status", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestUpdateFlightStatusNotFound(t *testing.T) {
	mux := newFlightsTestMux(store.New())

	body := strings.NewReader(`{"status":"cancelled"}`)
	req := httptest.NewRequest(http.MethodPatch, "/v1/flights/missing/status", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
