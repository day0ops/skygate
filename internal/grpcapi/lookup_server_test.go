package grpcapi

import (
	"context"
	"testing"

	"skygate/internal/domain"
	lookupv1 "skygate/internal/grpcapi/lookup/v1"
	"skygate/internal/store"
)

func TestLookupServerGetFlightStatus(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusAirborne, DelayMinutes: 5})

	s := NewLookupServer(st)
	resp, err := s.GetFlightStatus(context.Background(), &lookupv1.GetFlightStatusRequest{FlightId: "flt_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "airborne" || resp.DelayMinutes != 5 {
		t.Fatalf("got %+v, want status=airborne delayMinutes=5", resp)
	}
}

func TestLookupServerGetFlightStatusNotFound(t *testing.T) {
	s := NewLookupServer(store.New())
	if _, err := s.GetFlightStatus(context.Background(), &lookupv1.GetFlightStatusRequest{FlightId: "missing"}); err == nil {
		t.Fatal("expected error for missing flight, got nil")
	}
}
