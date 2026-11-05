// internal/store/store_test.go
package store

import (
	"testing"

	"skygate/internal/domain"
)

func TestFlightRoundTrip(t *testing.T) {
	s := New()
	f := domain.Flight{ID: "flt_1", FlightNumber: "SK100", Status: domain.StatusScheduled}
	s.PutFlight(f)

	got, ok := s.GetFlight("flt_1")
	if !ok {
		t.Fatal("expected flight to be found")
	}
	if got.FlightNumber != "SK100" {
		t.Fatalf("got flight number %q, want SK100", got.FlightNumber)
	}

	if len(s.ListFlights()) != 1 {
		t.Fatalf("got %d flights, want 1", len(s.ListFlights()))
	}

	s.DeleteFlight("flt_1")
	if _, ok := s.GetFlight("flt_1"); ok {
		t.Fatal("expected flight to be deleted")
	}
}

func TestAirportAndAircraftRoundTrip(t *testing.T) {
	s := New()
	s.PutAirport(domain.Airport{Code: "SFO", Name: "San Francisco Intl"})
	if _, ok := s.GetAirport("SFO"); !ok {
		t.Fatal("expected airport to be found")
	}
	if len(s.ListAirports()) != 1 {
		t.Fatalf("got %d airports, want 1", len(s.ListAirports()))
	}

	s.PutAircraft(domain.Aircraft{ID: "acf_1", TailNumber: "N100SK"})
	if _, ok := s.GetAircraft("acf_1"); !ok {
		t.Fatal("expected aircraft to be found")
	}
	if len(s.ListAircraft()) != 1 {
		t.Fatalf("got %d aircraft, want 1", len(s.ListAircraft()))
	}
}

func TestBookingRoundTrip(t *testing.T) {
	s := New()
	s.PutBooking(domain.Booking{ID: "bkg_1", FlightID: "flt_1"})
	if _, ok := s.GetBooking("bkg_1"); !ok {
		t.Fatal("expected booking to be found")
	}
	s.DeleteBooking("bkg_1")
	if _, ok := s.GetBooking("bkg_1"); ok {
		t.Fatal("expected booking to be deleted")
	}
}
