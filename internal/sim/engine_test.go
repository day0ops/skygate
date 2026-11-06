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

func TestAirborneDurationKeepsVelocityConstantAcrossRoutes(t *testing.T) {
	short := domain.Airport{Code: "A", X: 0, Y: 0}
	shortDest := domain.Airport{Code: "B", X: 67, Y: 0} // the real shortest pair in seedAirports (LHR-FRA).
	long := domain.Airport{Code: "C", X: 0, Y: 0}
	longDest := domain.Airport{Code: "D", X: 1441, Y: 0} // the real longest pair in seedAirports (SFO-SYD).

	shortDuration := airborneDurationFor(short, shortDest)
	longDuration := airborneDurationFor(long, longDest)

	shortVelocity := 67.0 / shortDuration.Seconds()
	longVelocity := 1441.0 / longDuration.Seconds()

	// A narrow duration clamp previously left routes up to 7.7x faster than
	// others (3.73 u/s vs 28.81 u/s) - velocity must now stay constant since
	// duration scales directly from distance with no clamp doing real work
	// across this range.
	if diff := math.Abs(shortVelocity - longVelocity); diff > 0.01 {
		t.Errorf(
			"expected velocity to stay constant across routes, got short=%.2f u/s long=%.2f u/s",
			shortVelocity, longVelocity,
		)
	}
}

func TestAirborneDurationClampsExtremeDistances(t *testing.T) {
	same := domain.Airport{Code: "A", X: 0, Y: 0}
	tiny := domain.Airport{Code: "B", X: 1, Y: 0}        // effectively zero distance.
	far := domain.Airport{Code: "C", X: 1_000_000, Y: 0} // far beyond any real airport pair.

	if d := airborneDurationFor(same, tiny); d != minAirborneDuration {
		t.Errorf("near-zero distance duration = %v, want the clamp floor %v", d, minAirborneDuration)
	}
	if d := airborneDurationFor(same, far); d != maxAirborneDuration {
		t.Errorf("extreme distance duration = %v, want the clamp ceiling %v", d, maxAirborneDuration)
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

func TestHoldingFlightHasPositionImmediatelyOnTransition(t *testing.T) {
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

	// A flight must never have a nil Position on the exact tick it enters
	// holding - the frontend falls back to the origin airport's position
	// whenever Position is nil, which previously made holding (and landing)
	// flights visibly snap backward across the map for one frame.
	if !changed || updated.Status != domain.StatusHolding {
		t.Fatalf("expected the flight to enter StatusHolding, got status=%v changed=%v", updated.Status, changed)
	}
	if updated.Position == nil {
		t.Fatal("expected Position to be set immediately on entering StatusHolding, got nil")
	}
	if updated.Position.X != dst.X+holdRadius || updated.Position.Y != dst.Y {
		t.Fatalf("expected the initial holding position at (dest.X+holdRadius, dest.Y), got %+v", updated.Position)
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
