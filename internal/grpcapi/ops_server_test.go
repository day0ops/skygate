package grpcapi

import (
	"context"
	"testing"
	"time"

	"skygate/internal/domain"
	opsv1 "skygate/internal/grpcapi/ops/v1"
	"skygate/internal/store"

	"google.golang.org/grpc"
)

func TestOpsServerAssignGate(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusBoarding})

	s := NewOpsServer(st)
	resp, err := s.AssignGate(context.Background(), &opsv1.AssignGateRequest{FlightId: "flt_1", Gate: "B12"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Gate != "B12" {
		t.Fatalf("got gate %q, want B12", resp.Gate)
	}

	updated, _ := st.GetFlight("flt_1")
	if updated.Gate != "B12" {
		t.Fatalf("store not updated: got gate %q", updated.Gate)
	}
}

func TestOpsServerAssignGateNotFound(t *testing.T) {
	s := NewOpsServer(store.New())
	if _, err := s.AssignGate(context.Background(), &opsv1.AssignGateRequest{FlightId: "missing"}); err == nil {
		t.Fatal("expected error for missing flight, got nil")
	}
}

func TestOpsServerUpdateFlightStatus(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusScheduled})

	s := NewOpsServer(st)
	resp, err := s.UpdateFlightStatus(context.Background(), &opsv1.UpdateFlightStatusRequest{FlightId: "flt_1", Status: "cancelled"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "cancelled" {
		t.Fatalf("got status %q, want cancelled", resp.Status)
	}
}

func TestOpsServerUpdateFlightStatusNotFound(t *testing.T) {
	s := NewOpsServer(store.New())
	if _, err := s.UpdateFlightStatus(context.Background(), &opsv1.UpdateFlightStatusRequest{FlightId: "missing", Status: "cancelled"}); err == nil {
		t.Fatal("expected error for missing flight, got nil")
	}
}

func TestOpsServerUpdateFlightStatusRejectsUnknownStatus(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusScheduled})

	s := NewOpsServer(st)
	if _, err := s.UpdateFlightStatus(context.Background(), &opsv1.UpdateFlightStatusRequest{FlightId: "flt_1", Status: "orbiting"}); err == nil {
		t.Fatal("expected error for unknown status, got nil")
	}

	f, _ := st.GetFlight("flt_1")
	if f.Status != domain.StatusScheduled {
		t.Fatalf("expected status to remain unchanged, got %q", f.Status)
	}
}

func TestOpsServerUpdateFlightStatusClearsPositionWhenLeavingAirborne(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{
		ID:       "flt_1",
		Status:   domain.StatusAirborne,
		Position: &domain.Position{X: 10, Y: 20, Progress: 0.5},
	})

	s := NewOpsServer(st)
	resp, err := s.UpdateFlightStatus(context.Background(), &opsv1.UpdateFlightStatusRequest{FlightId: "flt_1", Status: "landed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "landed" {
		t.Fatalf("got status %q, want landed", resp.Status)
	}

	f, _ := st.GetFlight("flt_1")
	if f.Position != nil {
		t.Fatalf("expected Position to be cleared after leaving airborne, got %+v", f.Position)
	}
}

func TestOpsServerUpdateFlightStatusPreservesPositionWhenHolding(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{
		ID:       "flt_1",
		Status:   domain.StatusAirborne,
		Position: &domain.Position{X: 10, Y: 20, Progress: 0.9},
	})

	s := NewOpsServer(st)
	resp, err := s.UpdateFlightStatus(context.Background(), &opsv1.UpdateFlightStatusRequest{FlightId: "flt_1", Status: "holding"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "holding" {
		t.Fatalf("got status %q, want holding", resp.Status)
	}

	f, _ := st.GetFlight("flt_1")
	if f.Position == nil {
		t.Fatal("expected Position to be preserved when transitioning to holding")
	}
}

type fakeStreamServer struct {
	grpc.ServerStream
	ctx  context.Context
	sent []*opsv1.Flight
}

func (f *fakeStreamServer) Context() context.Context { return f.ctx }

func (f *fakeStreamServer) Send(flight *opsv1.Flight) error {
	f.sent = append(f.sent, flight)
	return nil
}

func TestOpsServerStreamFleetStatus(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusAirborne})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	s := &OpsServer{store: st, streamInterval: 10 * time.Millisecond}
	fake := &fakeStreamServer{ctx: ctx}
	if err := s.StreamFleetStatus(&opsv1.Empty{}, fake); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.sent) == 0 {
		t.Fatal("expected at least one flight to be streamed")
	}
}
