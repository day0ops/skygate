package domain

import "time"

// FlightStatus is a Flight's position in its scheduled->landed lifecycle.
type FlightStatus string

const (
	StatusScheduled FlightStatus = "scheduled"
	StatusBoarding  FlightStatus = "boarding"
	StatusDeparted  FlightStatus = "departed"
	StatusAirborne  FlightStatus = "airborne"
	StatusLanded    FlightStatus = "landed"
	StatusDelayed   FlightStatus = "delayed"
	StatusCancelled FlightStatus = "cancelled"
)

// ValidFlightStatuses lists every known Flight.Status value. REST and gRPC
// write paths both validate against this one list so a flight can't end up
// with a status one API accepts and the other rejects.
var ValidFlightStatuses = []FlightStatus{
	StatusScheduled, StatusBoarding, StatusDeparted, StatusAirborne,
	StatusLanded, StatusDelayed, StatusCancelled,
}

// IsValidFlightStatus reports whether s matches one of ValidFlightStatuses.
func IsValidFlightStatus(s string) bool {
	for _, v := range ValidFlightStatuses {
		if string(v) == s {
			return true
		}
	}
	return false
}

// AircraftStatus tracks whether an Aircraft is free to be assigned to a
// flight.
type AircraftStatus string

const (
	AircraftAvailable   AircraftStatus = "available"
	AircraftInFlight    AircraftStatus = "in-flight"
	AircraftMaintenance AircraftStatus = "maintenance"
)

// BookingClass is the cabin class a Booking reserves.
type BookingClass string

const (
	ClassEconomy  BookingClass = "economy"
	ClassBusiness BookingClass = "business"
	ClassFirst    BookingClass = "first"
)

// BookingStatus is a Booking's current state.
type BookingStatus string

const (
	BookingConfirmed BookingStatus = "confirmed"
	BookingCheckedIn BookingStatus = "checked-in"
	BookingCancelled BookingStatus = "cancelled"
)

// Position is a Flight's location on the schematic map while airborne.
type Position struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Heading  float64 `json:"heading"`
	Progress float64 `json:"progress"`
}

// Gate is a single boarding gate at an Airport.
type Gate struct {
	Code string `json:"code"`
}

// Airport is one of the fixed, seeded locations flights depart from and
// arrive at.
type Airport struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	City  string  `json:"city"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Gates []Gate  `json:"gates"`
}

// Aircraft is a single tail in the simulated fleet, assignable to one
// Flight at a time.
type Aircraft struct {
	ID              string         `json:"id"`
	TailNumber      string         `json:"tailNumber"`
	Model           string         `json:"model"`
	Capacity        int            `json:"capacity"`
	Status          AircraftStatus `json:"status"`
	CurrentFlightID string         `json:"currentFlightId,omitempty"`
}

// Flight is a single simulated flight moving through its lifecycle.
type Flight struct {
	ID                 string       `json:"id"`
	FlightNumber       string       `json:"flightNumber"`
	OriginCode         string       `json:"originCode"`
	DestinationCode    string       `json:"destinationCode"`
	AircraftID         string       `json:"aircraftId"`
	Status             FlightStatus `json:"status"`
	ScheduledDeparture time.Time    `json:"scheduledDeparture"`
	ScheduledArrival   time.Time    `json:"scheduledArrival"`
	Gate               string       `json:"gate,omitempty"`
	Position           *Position    `json:"position,omitempty"`
	DelayMinutes       int          `json:"delayMinutes,omitempty"`
}

// Booking reserves a seat for a passenger on a Flight.
type Booking struct {
	ID            string        `json:"id"`
	FlightID      string        `json:"flightId"`
	PassengerName string        `json:"passengerName"`
	Seat          string        `json:"seat"`
	Class         BookingClass  `json:"class"`
	Status        BookingStatus `json:"status"`
}
