package grpcapi

import (
	"context"

	lookupv1 "skygate/internal/grpcapi/lookup/v1"
	"skygate/internal/store"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LookupServer implements lookupv1.FlightLookupServiceServer against st.
// kgateway transcodes GetFlightStatus into REST/JSON using the proto's
// google.api.http annotation - there is no native REST handler for it
// anywhere in this codebase.
type LookupServer struct {
	lookupv1.UnimplementedFlightLookupServiceServer
	store *store.Store
}

// NewLookupServer returns a LookupServer backed by st.
func NewLookupServer(st *store.Store) *LookupServer {
	return &LookupServer{store: st}
}

// GetFlightStatus returns the status and delay for a flight, or NotFound if the flight doesn't exist.
func (s *LookupServer) GetFlightStatus(ctx context.Context, req *lookupv1.GetFlightStatusRequest) (*lookupv1.FlightStatusResponse, error) {
	f, ok := s.store.GetFlight(req.GetFlightId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "flight %q not found", req.GetFlightId())
	}
	return &lookupv1.FlightStatusResponse{
		FlightId:     f.ID,
		Status:       string(f.Status),
		DelayMinutes: int32(f.DelayMinutes),
	}, nil
}
