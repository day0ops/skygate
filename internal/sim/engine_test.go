package sim

import (
	"testing"
	"time"

	"skygate/internal/domain"
	"skygate/internal/store"
)

func TestEngineSpawnsAndAdvances(t *testing.T) {
	st := store.New()
	clock := NewManualClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	e := NewEngine(st, clock, 1)
	e.Seed()

	e.Tick()
	flights := st.ListFlights()
	if len(flights) != 1 {
		t.Fatalf("got %d flights after first tick, want 1", len(flights))
	}
	firstID := flights[0].ID
	if flights[0].Status != domain.StatusScheduled && flights[0].Status != domain.StatusDelayed {
		t.Fatalf("got status %q, want scheduled or delayed", flights[0].Status)
	}

	// Advance well past the scheduled+delay window and tick again. The
	// spawn interval (6s) is shorter than this 40s jump, so more flights
	// may have spawned by now too - this only asserts on the original one.
	clock.Advance(40 * time.Second)
	e.Tick()
	f, ok := st.GetFlight(firstID)
	if !ok {
		t.Fatalf("flight %q disappeared", firstID)
	}
	if f.Status != domain.StatusBoarding {
		t.Fatalf("got status %q, want boarding", f.Status)
	}
	if f.Gate == "" {
		t.Fatal("expected a gate to be assigned once boarding")
	}
}

func TestEngineSubscribeIsNotifiedOnSpawn(t *testing.T) {
	st := store.New()
	clock := NewManualClock(time.Now())
	e := NewEngine(st, clock, 2)
	e.Seed()

	var seen []domain.Flight
	e.Subscribe(func(f domain.Flight) { seen = append(seen, f) })

	e.Tick()
	if len(seen) != 1 {
		t.Fatalf("got %d notifications, want 1", len(seen))
	}
}

func TestEngineInterpolatesAirbornePosition(t *testing.T) {
	st := store.New()
	clock := NewManualClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	e := NewEngine(st, clock, 3)
	e.Seed()

	e.Tick() // spawn
	// Advance through scheduled, boarding, and departed to reach airborne.
	clock.Advance(scheduledDuration)
	e.Tick() // scheduled -> boarding
	clock.Advance(boardingDuration)
	e.Tick() // boarding -> departed
	clock.Advance(departedDuration)
	e.Tick() // departed -> airborne
	e.Tick() // airborne: position gets interpolated

	flights := st.ListFlights()
	var f domain.Flight
	found := false
	for _, candidate := range flights {
		if candidate.Status == domain.StatusAirborne {
			f = candidate
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no airborne flight found")
	}
	if f.Position == nil {
		t.Fatal("expected a position once airborne")
	}
	if f.Position.Progress < 0 || f.Position.Progress > 1 {
		t.Fatalf("got progress %f, want between 0 and 1", f.Position.Progress)
	}

	// Halfway through the airborne window, progress should have advanced.
	clock.Advance(airborneDuration / 2)
	e.Tick()
	f, _ = st.GetFlight(f.ID)
	if f.Position == nil || f.Position.Progress < 0.3 {
		t.Fatalf("expected progress to have advanced, got %+v", f.Position)
	}
}
