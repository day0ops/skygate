package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"skygate/internal/domain"
	"skygate/internal/store"
)

func TestRegisterAirports(t *testing.T) {
	st := store.New()
	st.PutAirport(domain.Airport{Code: "SFO", Name: "San Francisco Intl", City: "San Francisco"})

	mux := http.NewServeMux()
	RegisterAirports(mux, st)

	t.Run("list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/airports", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want 200", rec.Code)
		}
	})

	t.Run("detail found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/airports/SFO", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want 200", rec.Code)
		}
	})

	t.Run("detail not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/airports/ZZZ", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("got status %d, want 404", rec.Code)
		}
	})
}
