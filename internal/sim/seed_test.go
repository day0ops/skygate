package sim

import (
	"testing"

	"skygate/internal/domain"
)

func TestSeedAirportsAreUnique(t *testing.T) {
	airports := seedAirports()
	if len(airports) < 2 {
		t.Fatalf("expected at least 2 airports, got %d", len(airports))
	}
	seen := make(map[string]bool)
	for _, a := range airports {
		if seen[a.Code] {
			t.Fatalf("duplicate airport code %q", a.Code)
		}
		seen[a.Code] = true
	}
}

func TestSeedAircraftAreAvailable(t *testing.T) {
	aircraft := seedAircraft()
	if len(aircraft) == 0 {
		t.Fatal("expected at least one aircraft")
	}
	for _, a := range aircraft {
		if a.Status != domain.AircraftAvailable {
			t.Fatalf("aircraft %q: got status %q, want %q", a.ID, a.Status, domain.AircraftAvailable)
		}
		if a.ID == "" {
			t.Fatal("expected aircraft to have an ID")
		}
	}
}
