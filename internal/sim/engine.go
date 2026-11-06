package sim

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"skygate/internal/domain"
	"skygate/internal/store"
)

// Lifecycle stage durations, compressed so a full scheduled->landed cycle
// takes roughly a minute - enough to be visibly dynamic in a live demo.
const (
	scheduledDuration = 10 * time.Second
	boardingDuration  = 8 * time.Second
	departedDuration  = 4 * time.Second
	landedDuration    = 8 * time.Second

	// Airborne duration scales with actual distance between airports instead
	// of being fixed, clamped so neither the shortest nor longest route in
	// seedAirports breaks the "visibly dynamic in a live demo" pacing above.
	referenceDistance        = 700.0
	referenceAirborneSeconds = 30.0
	minAirborneDuration      = 18 * time.Second
	maxAirborneDuration      = 50 * time.Second

	maxActiveFlights = 12
	spawnInterval    = 6 * time.Second
	delayChance      = 0.2
)

// airborneDurationFor scales flight duration by the actual distance between
// origin and dest, clamped to [minAirborneDuration, maxAirborneDuration] so
// routes stay demo-paced at both extremes while still feeling different.
func airborneDurationFor(origin, dest domain.Airport) time.Duration {
	dist := math.Hypot(dest.X-origin.X, dest.Y-origin.Y)
	seconds := referenceAirborneSeconds * (dist / referenceDistance)
	d := time.Duration(seconds * float64(time.Second))
	if d < minAirborneDuration {
		return minAirborneDuration
	}
	if d > maxAirborneDuration {
		return maxAirborneDuration
	}
	return d
}

// Engine advances the simulated flight world forward in time and notifies
// subscribers whenever a flight is created or changes.
type Engine struct {
	store *store.Store
	clock Clock
	rng   *rand.Rand

	mu         sync.Mutex
	stateSince map[string]time.Time
	nextSpawn  time.Time

	listenersMu sync.RWMutex
	listeners   []func(domain.Flight)
}

// NewEngine builds an Engine backed by st, using clock for all timing
// decisions so tests can drive it deterministically.
func NewEngine(st *store.Store, clock Clock, seed int64) *Engine {
	return &Engine{
		store:      st,
		clock:      clock,
		rng:        rand.New(rand.NewSource(seed)),
		stateSince: make(map[string]time.Time),
	}
}

// Seed populates the store with the fixed airports/aircraft pool. Call once
// at startup before the first Tick.
func (e *Engine) Seed() {
	for _, a := range seedAirports() {
		e.store.PutAirport(a)
	}
	for _, a := range seedAircraft() {
		e.store.PutAircraft(a)
	}
	e.nextSpawn = e.clock.Now()
}

// Subscribe registers fn to be called whenever a flight is created or
// changes state. fn is called synchronously from Tick while e.mu is held,
// so it must not block and must not call back into the Engine (Tick,
// Seed, Subscribe) or it will deadlock on the non-reentrant mutex.
func (e *Engine) Subscribe(fn func(domain.Flight)) {
	e.listenersMu.Lock()
	defer e.listenersMu.Unlock()
	e.listeners = append(e.listeners, fn)
}

func (e *Engine) notify(f domain.Flight) {
	e.listenersMu.RLock()
	defer e.listenersMu.RUnlock()
	for _, fn := range e.listeners {
		fn(f)
	}
}

// Tick advances the simulation by one step: it may spawn a new flight and
// advances every in-progress flight's lifecycle based on elapsed time.
func (e *Engine) Tick() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.clock.Now()
	if !now.Before(e.nextSpawn) {
		e.spawnFlight(now)
		e.nextSpawn = now.Add(spawnInterval)
	}

	for _, f := range e.store.ListFlights() {
		updated, changed, remove := e.advance(f, now)
		if remove {
			e.store.DeleteFlight(updated.ID)
			delete(e.stateSince, updated.ID)
			continue
		}
		if changed {
			e.store.PutFlight(updated)
			e.notify(updated)
		}
	}
}

// Run calls Tick every interval until ctx is cancelled.
func (e *Engine) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			e.Tick()
		}
	}
}

func (e *Engine) spawnFlight(now time.Time) {
	active := 0
	for _, f := range e.store.ListFlights() {
		if f.Status != domain.StatusLanded && f.Status != domain.StatusCancelled {
			active++
		}
	}
	if active >= maxActiveFlights {
		return
	}

	airports := e.store.ListAirports()
	if len(airports) < 2 {
		return
	}
	origin := airports[e.rng.Intn(len(airports))]
	dest := origin
	for dest.Code == origin.Code {
		dest = airports[e.rng.Intn(len(airports))]
	}

	aircraft, ok := e.availableAircraft()
	if !ok {
		return
	}

	status := domain.StatusScheduled
	delay := 0
	if e.rng.Float64() < delayChance {
		status = domain.StatusDelayed
		delay = 5 + e.rng.Intn(25)
	}

	f := domain.Flight{
		ID:                 domain.NewID("flt"),
		FlightNumber:       fmt.Sprintf("SK%d", 1000+e.rng.Intn(9000)),
		OriginCode:         origin.Code,
		DestinationCode:    dest.Code,
		AircraftID:         aircraft.ID,
		Status:             status,
		DelayMinutes:       delay,
		ScheduledDeparture: now.Add(scheduledDuration + boardingDuration),
		ScheduledArrival:   now.Add(scheduledDuration + boardingDuration + departedDuration + airborneDurationFor(origin, dest)),
	}

	aircraft.Status = domain.AircraftInFlight
	aircraft.CurrentFlightID = f.ID
	e.store.PutAircraft(aircraft)

	e.store.PutFlight(f)
	e.stateSince[f.ID] = now
	e.notify(f)
}

func (e *Engine) availableAircraft() (domain.Aircraft, bool) {
	for _, a := range e.store.ListAircraft() {
		if a.Status == domain.AircraftAvailable {
			return a, true
		}
	}
	return domain.Aircraft{}, false
}

// advance returns f transitioned forward if it has spent long enough in its
// current status. changed reports whether f was mutated; remove reports
// whether the flight should be deleted from the store (its demo lifecycle
// is complete).
func (e *Engine) advance(f domain.Flight, now time.Time) (updated domain.Flight, changed bool, remove bool) {
	since, ok := e.stateSince[f.ID]
	if !ok {
		since = now
		e.stateSince[f.ID] = since
	}
	elapsed := now.Sub(since)

	switch f.Status {
	case domain.StatusDelayed:
		if elapsed >= scheduledDuration+time.Duration(f.DelayMinutes)*time.Second {
			f.Status = domain.StatusBoarding
			f.Gate = e.randomGate()
			f.DelayMinutes = 0
			e.stateSince[f.ID] = now
			return f, true, false
		}
	case domain.StatusScheduled:
		if elapsed >= scheduledDuration {
			f.Status = domain.StatusBoarding
			f.Gate = e.randomGate()
			e.stateSince[f.ID] = now
			return f, true, false
		}
	case domain.StatusBoarding:
		if elapsed >= boardingDuration {
			f.Status = domain.StatusDeparted
			e.stateSince[f.ID] = now
			return f, true, false
		}
	case domain.StatusDeparted:
		if elapsed >= departedDuration {
			f.Status = domain.StatusAirborne
			e.stateSince[f.ID] = now
			return f, true, false
		}
	case domain.StatusAirborne:
		origin, _ := e.store.GetAirport(f.OriginCode)
		dest, _ := e.store.GetAirport(f.DestinationCode)
		progress := float64(elapsed) / float64(airborneDurationFor(origin, dest))
		if progress >= 1 {
			f.Status = domain.StatusLanded
			f.Position = nil
			e.releaseAircraft(f.AircraftID)
			e.stateSince[f.ID] = now
			return f, true, false
		}
		e.updatePosition(&f, progress)
		return f, true, false
	case domain.StatusLanded:
		if elapsed >= landedDuration {
			return f, false, true
		}
	}
	return f, false, false
}

func (e *Engine) randomGate() string {
	return fmt.Sprintf("%c%d", 'A'+e.rng.Intn(6), 1+e.rng.Intn(9))
}

func (e *Engine) releaseAircraft(id string) {
	a, ok := e.store.GetAircraft(id)
	if !ok {
		return
	}
	a.Status = domain.AircraftAvailable
	a.CurrentFlightID = ""
	e.store.PutAircraft(a)
}

func (e *Engine) updatePosition(f *domain.Flight, progress float64) {
	origin, _ := e.store.GetAirport(f.OriginCode)
	dest, _ := e.store.GetAirport(f.DestinationCode)
	x := origin.X + (dest.X-origin.X)*progress
	y := origin.Y + (dest.Y-origin.Y)*progress
	// atan2 measures from due East; the plane icon's native orientation points
	// North (up) at rotate(0), so the raw angle needs a 90-degree offset to
	// actually face the travel direction once the frontend rotates it.
	heading := math.Atan2(dest.Y-origin.Y, dest.X-origin.X)*180/math.Pi + 90
	f.Position = &domain.Position{X: x, Y: y, Heading: heading, Progress: progress}
}
