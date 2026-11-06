package rest

import (
	"encoding/json"
	"net/http"

	"skygate/internal/domain"
	"skygate/internal/store"
)

// RegisterFlights adds the flight read and write endpoints to mux.
func RegisterFlights(mux *http.ServeMux, st *store.Store) {
	mux.HandleFunc("GET /v1/flights", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		status := q.Get("status")
		origin := q.Get("origin")
		destination := q.Get("destination")

		flights := st.ListFlights()
		filtered := make([]domain.Flight, 0, len(flights))
		for _, f := range flights {
			if status != "" && string(f.Status) != status {
				continue
			}
			if origin != "" && f.OriginCode != origin {
				continue
			}
			if destination != "" && f.DestinationCode != destination {
				continue
			}
			filtered = append(filtered, f)
		}
		writeJSON(w, http.StatusOK, filtered)
	})

	mux.HandleFunc("GET /v1/flights/{id}", func(w http.ResponseWriter, r *http.Request) {
		f, ok := st.GetFlight(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "flight not found")
			return
		}
		writeJSON(w, http.StatusOK, f)
	})

	mux.HandleFunc("GET /v1/flights/{id}/position", func(w http.ResponseWriter, r *http.Request) {
		f, ok := st.GetFlight(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "flight not found")
			return
		}
		if f.Position == nil {
			writeError(w, http.StatusConflict, "flight is not airborne")
			return
		}
		writeJSON(w, http.StatusOK, f.Position)
	})

	mux.HandleFunc("POST /v1/flights", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			FlightNumber    string `json:"flightNumber"`
			OriginCode      string `json:"originCode"`
			DestinationCode string `json:"destinationCode"`
			AircraftID      string `json:"aircraftId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.FlightNumber == "" || req.OriginCode == "" || req.DestinationCode == "" {
			writeError(w, http.StatusBadRequest, "flightNumber, originCode and destinationCode are required")
			return
		}
		f := domain.Flight{
			ID:              domain.NewID("flt"),
			FlightNumber:    req.FlightNumber,
			OriginCode:      req.OriginCode,
			DestinationCode: req.DestinationCode,
			AircraftID:      req.AircraftID,
			Status:          domain.StatusScheduled,
		}
		st.PutFlight(f)
		writeJSON(w, http.StatusCreated, f)
	})

	mux.HandleFunc("PATCH /v1/flights/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		f, ok := st.GetFlight(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "flight not found")
			return
		}
		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !domain.IsValidFlightStatus(req.Status) {
			writeError(w, http.StatusBadRequest, "status must be one of the known flight statuses")
			return
		}
		f.Status = domain.FlightStatus(req.Status)
		if f.Status != domain.StatusAirborne && f.Status != domain.StatusHolding {
			f.Position = nil
		}
		st.PutFlight(f)
		writeJSON(w, http.StatusOK, f)
	})
}
