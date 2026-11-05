package sim

import (
	"fmt"

	"skygate/internal/domain"
)

// seedAirports returns the fixed set of airports skygate operates between.
// Coordinates are an abstract 1600x650 schematic space, not real geography -
// the UI draws a stylized map, not real map tiles.
func seedAirports() []domain.Airport {
	return []domain.Airport{
		{Code: "SFO", Name: "San Francisco Intl", City: "San Francisco", X: 50, Y: 220, Gates: []domain.Gate{{Code: "A1"}, {Code: "A2"}}},
		{Code: "JFK", Name: "John F. Kennedy Intl", City: "New York", X: 420, Y: 180, Gates: []domain.Gate{{Code: "B1"}, {Code: "B2"}}},
		{Code: "LHR", Name: "London Heathrow", City: "London", X: 700, Y: 120, Gates: []domain.Gate{{Code: "C1"}, {Code: "C2"}}},
		{Code: "FRA", Name: "Frankfurt am Main", City: "Frankfurt", X: 760, Y: 150, Gates: []domain.Gate{{Code: "D1"}, {Code: "D2"}}},
		{Code: "DXB", Name: "Dubai Intl", City: "Dubai", X: 980, Y: 300, Gates: []domain.Gate{{Code: "E1"}, {Code: "E2"}}},
		{Code: "SIN", Name: "Singapore Changi", City: "Singapore", X: 1250, Y: 400, Gates: []domain.Gate{{Code: "F1"}, {Code: "F2"}}},
		{Code: "NRT", Name: "Narita Intl", City: "Tokyo", X: 1400, Y: 250, Gates: []domain.Gate{{Code: "G1"}, {Code: "G2"}}},
		{Code: "SYD", Name: "Sydney Kingsford Smith", City: "Sydney", X: 1450, Y: 560, Gates: []domain.Gate{{Code: "H1"}, {Code: "H2"}}},
		{Code: "GRU", Name: "São Paulo/Guarulhos", City: "São Paulo", X: 430, Y: 560, Gates: []domain.Gate{{Code: "I1"}, {Code: "I2"}}},
		{Code: "ORD", Name: "Chicago O'Hare", City: "Chicago", X: 300, Y: 170, Gates: []domain.Gate{{Code: "J1"}, {Code: "J2"}}},
	}
}

// seedAircraft returns the fixed pool of aircraft skygate schedules flights
// with.
func seedAircraft() []domain.Aircraft {
	models := []string{"Boeing 737-800", "Airbus A320", "Boeing 787-9", "Airbus A350"}
	aircraft := make([]domain.Aircraft, 0, 8)
	for i := 0; i < 8; i++ {
		aircraft = append(aircraft, domain.Aircraft{
			ID:         domain.NewID("acf"),
			TailNumber: fmt.Sprintf("N%d%cSK", 100+i, 'A'+i),
			Model:      models[i%len(models)],
			Capacity:   160 + (i%3)*40,
			Status:     domain.AircraftAvailable,
		})
	}
	return aircraft
}
