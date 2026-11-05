package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"skygate/internal/domain"
	"skygate/internal/store"
)

func newBookingsTestMux(st *store.Store) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterBookings(mux, st)
	return mux
}

func TestCreateBooking(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1"})
	mux := newBookingsTestMux(st)

	body := strings.NewReader(`{"flightId":"flt_1","passengerName":"Ada Lovelace","seat":"12A","class":"economy"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/bookings", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, want 201, body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateBookingRejectsUnknownFlight(t *testing.T) {
	mux := newBookingsTestMux(store.New())
	body := strings.NewReader(`{"flightId":"missing","passengerName":"Ada","seat":"12A","class":"economy"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/bookings", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestCreateBookingRejectsInvalidBody(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1"})
	mux := newBookingsTestMux(st)

	body := strings.NewReader(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/v1/bookings", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestCreateBookingRejectsMissingFields(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1"})
	mux := newBookingsTestMux(st)

	body := strings.NewReader(`{"flightId":"flt_1","passengerName":"","seat":"12A","class":"economy"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/bookings", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestCreateBookingRejectsInvalidClass(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1"})
	mux := newBookingsTestMux(st)

	body := strings.NewReader(`{"flightId":"flt_1","passengerName":"Ada","seat":"12A","class":"invalid"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/bookings", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", rec.Code)
	}
}

func TestGetBookingNotFound(t *testing.T) {
	mux := newBookingsTestMux(store.New())

	req := httptest.NewRequest(http.MethodGet, "/v1/bookings/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestDeleteBookingNotFound(t *testing.T) {
	mux := newBookingsTestMux(store.New())

	req := httptest.NewRequest(http.MethodDelete, "/v1/bookings/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestGetAndDeleteBooking(t *testing.T) {
	st := store.New()
	st.PutBooking(domain.Booking{ID: "bkg_1", FlightID: "flt_1"})
	mux := newBookingsTestMux(st)

	req := httptest.NewRequest(http.MethodGet, "/v1/bookings/bkg_1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rec.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/v1/bookings/bkg_1", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want 204", rec.Code)
	}

	if _, ok := st.GetBooking("bkg_1"); ok {
		t.Fatal("expected booking to be deleted")
	}
}
