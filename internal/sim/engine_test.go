package sim

import (
	"fmt"
	"math"
	"testing"
	"time"

	"skygate/internal/domain"
	"skygate/internal/store"
)

func TestUpdatePositionIsDeterministicPerFlightID(t *testing.T) {
	st := store.New()
	st.PutAirport(domain.Airport{Code: "ORI", X: 0, Y: 0})
	st.PutAirport(domain.Airport{Code: "DST", X: 100, Y: 0})
	e := NewEngine(st, RealClock, 1)

	f1 := domain.Flight{ID: "flt_test_position", OriginCode: "ORI", DestinationCode: "DST"}
	e.updatePosition(&f1, 0.5)
	if f1.Position == nil {
		t.Fatal("expected a position")
	}
	if math.IsNaN(f1.Position.Heading) || math.IsInf(f1.Position.Heading, 0) {
		t.Fatalf("expected a finite heading, got %v", f1.Position.Heading)
	}

	// Calling again with the same flight ID must produce the exact same
	// position - the curve shape is derived from the ID and must be stable.
	f2 := domain.Flight{ID: "flt_test_position", OriginCode: "ORI", DestinationCode: "DST"}
	e.updatePosition(&f2, 0.5)
	if f1.Position.X != f2.Position.X || f1.Position.Y != f2.Position.Y || f1.Position.Heading != f2.Position.Heading {
		t.Fatalf("expected the same flight ID to produce the same position, got %+v and %+v", f1.Position, f2.Position)
	}
}

func TestAirborneDurationScalesWithDistanceAndClamps(t *testing.T) {
	short := domain.Airport{Code: "A", X: 0, Y: 0}
	shortDest := domain.Airport{Code: "B", X: 100, Y: 0} // well under the clamp floor's reference distance.
	long := domain.Airport{Code: "C", X: 0, Y: 0}
	longDest := domain.Airport{Code: "D", X: 5000, Y: 0} // well over the clamp ceiling's reference distance.
	mid := domain.Airport{Code: "E", X: 0, Y: 0}
	midDest := domain.Airport{Code: "F", X: 700, Y: 0} // exactly the reference distance.

	shortDuration := airborneDurationFor(short, shortDest)
	longDuration := airborneDurationFor(long, longDest)
	midDuration := airborneDurationFor(mid, midDest)

	if shortDuration != minAirborneDuration {
		t.Errorf("short route duration = %v, want the clamp floor %v", shortDuration, minAirborneDuration)
	}
	if longDuration != maxAirborneDuration {
		t.Errorf("long route duration = %v, want the clamp ceiling %v", longDuration, maxAirborneDuration)
	}
	if midDuration != 30*time.Second {
		t.Errorf("reference-distance route duration = %v, want 30s", midDuration)
	}
	if shortDuration >= midDuration || midDuration >= longDuration {
		t.Errorf(
			"expected duration to increase with distance, got short=%v mid=%v long=%v",
			shortDuration, midDuration, longDuration,
		)
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
	origin, _ := st.GetAirport(f.OriginCode)
	dest, _ := st.GetAirport(f.DestinationCode)
	clock.Advance(airborneDurationFor(origin, dest) / 2)
	e.Tick()
	f, _ = st.GetFlight(f.ID)
	if f.Position == nil || f.Position.Progress < 0.3 {
		t.Fatalf("expected progress to have advanced, got %+v", f.Position)
	}
}

func TestHoldingFlightLandsAfterHoldingDuration(t *testing.T) {
	st := store.New()
	st.PutAirport(domain.Airport{Code: "ORI", X: 0, Y: 0})
	st.PutAirport(domain.Airport{Code: "DST", X: 100, Y: 0})
	clock := NewManualClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	e := NewEngine(st, clock, 1)

	// Find a flight ID that holdDecision says will hold, so this test is
	// deterministic rather than depending on luck.
	var holdingID string
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("flt_search_%d", i)
		if holds, _ := holdDecision(id); holds {
			holdingID = id
			break
		}
	}
	if holdingID == "" {
		t.Fatal("expected to find at least one flight ID that holds within 1000 tries")
	}

	aircraft := domain.Aircraft{ID: "acf_test", Status: domain.AircraftAvailable}
	st.PutAircraft(aircraft)
	f := domain.Flight{
		ID: holdingID, OriginCode: "ORI", DestinationCode: "DST", AircraftID: "acf_test", Status: domain.StatusAirborne,
	}
	st.PutFlight(f)

	ori, _ := st.GetAirport("ORI")
	dst, _ := st.GetAirport("DST")
	e.stateSince[f.ID] = clock.Now()
	updated, changed, _ := e.advance(f, clock.Now().Add(airborneDurationFor(ori, dst)))
	if !changed || updated.Status != domain.StatusHolding {
		t.Fatalf("expected the flight to enter StatusHolding, got status=%v changed=%v", updated.Status, changed)
	}

	e.stateSince[updated.ID] = clock.Now()
	landed, changed, _ := e.advance(updated, clock.Now().Add(holdingDuration))
	if !changed || landed.Status != domain.StatusLanded {
		t.Fatalf("expected the flight to land after holdingDuration, got status=%v changed=%v", landed.Status, changed)
	}
}
