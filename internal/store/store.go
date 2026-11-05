package store

import (
	"sync"

	"skygate/internal/domain"
)

// Store is a thread-safe in-memory holder for all skygate domain data.
// State is intentionally volatile: a restart resets the demo to a clean slate.
//
// Flight.Position and Airport.Gates are reference types, so Get/List return
// values that alias store-internal memory through those fields. Callers
// must never mutate through them - always build an updated copy and pass it
// to the matching Put method instead.
type Store struct {
	mu       sync.RWMutex
	airports map[string]domain.Airport
	aircraft map[string]domain.Aircraft
	flights  map[string]domain.Flight
	bookings map[string]domain.Booking
}

// New returns an empty Store.
func New() *Store {
	return &Store{
		airports: make(map[string]domain.Airport),
		aircraft: make(map[string]domain.Aircraft),
		flights:  make(map[string]domain.Flight),
		bookings: make(map[string]domain.Booking),
	}
}

// PutAirport inserts or replaces the airport keyed by its Code.
func (s *Store) PutAirport(a domain.Airport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.airports[a.Code] = a
}

// GetAirport returns the airport with the given code, if present.
func (s *Store) GetAirport(code string) (domain.Airport, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.airports[code]
	return a, ok
}

// ListAirports returns every airport in the store, in no particular order.
func (s *Store) ListAirports() []domain.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Airport, 0, len(s.airports))
	for _, a := range s.airports {
		out = append(out, a)
	}
	return out
}

// PutAircraft inserts or replaces the aircraft keyed by its ID.
func (s *Store) PutAircraft(a domain.Aircraft) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aircraft[a.ID] = a
}

// GetAircraft returns the aircraft with the given ID, if present.
func (s *Store) GetAircraft(id string) (domain.Aircraft, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.aircraft[id]
	return a, ok
}

// ListAircraft returns every aircraft in the store, in no particular order.
func (s *Store) ListAircraft() []domain.Aircraft {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Aircraft, 0, len(s.aircraft))
	for _, a := range s.aircraft {
		out = append(out, a)
	}
	return out
}

// PutFlight inserts or replaces the flight keyed by its ID.
func (s *Store) PutFlight(f domain.Flight) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flights[f.ID] = f
}

// GetFlight returns the flight with the given ID, if present.
func (s *Store) GetFlight(id string) (domain.Flight, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flights[id]
	return f, ok
}

// ListFlights returns every flight in the store, in no particular order.
func (s *Store) ListFlights() []domain.Flight {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Flight, 0, len(s.flights))
	for _, f := range s.flights {
		out = append(out, f)
	}
	return out
}

// DeleteFlight removes the flight with the given ID, if present.
func (s *Store) DeleteFlight(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.flights, id)
}

// PutBooking inserts or replaces the booking keyed by its ID.
func (s *Store) PutBooking(b domain.Booking) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bookings[b.ID] = b
}

// GetBooking returns the booking with the given ID, if present.
func (s *Store) GetBooking(id string) (domain.Booking, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.bookings[id]
	return b, ok
}

// DeleteBooking removes the booking with the given ID, if present.
func (s *Store) DeleteBooking(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.bookings, id)
}
