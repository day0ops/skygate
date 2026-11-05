package rest

import (
	"encoding/json"
	"net/http"

	"skygate/internal/domain"
	"skygate/internal/store"
)

var validBookingClasses = map[string]bool{
	string(domain.ClassEconomy):  true,
	string(domain.ClassBusiness): true,
	string(domain.ClassFirst):    true,
}

// RegisterBookings adds booking endpoints to mux.
func RegisterBookings(mux *http.ServeMux, st *store.Store) {
	mux.HandleFunc("POST /v1/bookings", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			FlightID      string `json:"flightId"`
			PassengerName string `json:"passengerName"`
			Seat          string `json:"seat"`
			Class         string `json:"class"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if _, ok := st.GetFlight(req.FlightID); !ok {
			writeError(w, http.StatusBadRequest, "flightId does not exist")
			return
		}
		if req.PassengerName == "" || req.Seat == "" || !validBookingClasses[req.Class] {
			writeError(w, http.StatusBadRequest, "passengerName, seat and a valid class are required")
			return
		}
		b := domain.Booking{
			ID:            domain.NewID("bkg"),
			FlightID:      req.FlightID,
			PassengerName: req.PassengerName,
			Seat:          req.Seat,
			Class:         domain.BookingClass(req.Class),
			Status:        domain.BookingConfirmed,
		}
		st.PutBooking(b)
		writeJSON(w, http.StatusCreated, b)
	})

	mux.HandleFunc("GET /v1/bookings/{id}", func(w http.ResponseWriter, r *http.Request) {
		b, ok := st.GetBooking(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "booking not found")
			return
		}
		writeJSON(w, http.StatusOK, b)
	})

	mux.HandleFunc("DELETE /v1/bookings/{id}", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := st.GetBooking(r.PathValue("id")); !ok {
			writeError(w, http.StatusNotFound, "booking not found")
			return
		}
		st.DeleteBooking(r.PathValue("id"))
		w.WriteHeader(http.StatusNoContent)
	})
}
