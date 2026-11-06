package sim

import (
	"math"
	"testing"
	"time"

	"skygate/internal/domain"
	"skygate/internal/store"
)

// TestUpdatePositionHeadingMatchesTravelDirection covers the plane icon's
// native orientation, which points up/North at rotate(0) - the frontend
// rotates it by Position.Heading to face the direction of travel, so the
// heading math must account for that native orientation, not just the raw
// atan2 angle (which is measured from due East).
func TestUpdatePositionHeadingMatchesTravelDirection(t *testing.T) {
	st := store.New()
	e := NewEngine(st, RealClock, 1)

	cases := []struct {
		name        string
		dx, dy      float64
		wantHeading float64
	}{
		{"due east", 10, 0, 90},
		{"due south", 0, 10, 180},
		{"due north", 0, -10, 0},
		{"due west", -10, 0, 270},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st.PutAirport(domain.Airport{Code: "ORI", X: 0, Y: 0})
			st.PutAirport(domain.Airport{Code: "DST", X: tc.dx, Y: tc.dy})

			f := domain.Flight{OriginCode: "ORI", DestinationCode: "DST"}
			e.updatePosition(&f, 0.5)

			if f.Position == nil {
				t.Fatal("expected a position")
			}
			if diff := math.Abs(f.Position.Heading - tc.wantHeading); diff > 0.001 {
				t.Fatalf("heading = %v, want %v", f.Position.Heading, tc.wantHeading)
			}
		})
	}
}

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
