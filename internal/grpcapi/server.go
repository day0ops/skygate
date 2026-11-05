package grpcapi

import (
	"google.golang.org/grpc"

	lookupv1 "skygate/internal/grpcapi/lookup/v1"
	opsv1 "skygate/internal/grpcapi/ops/v1"
	"skygate/internal/store"
)

// NewServer builds a *grpc.Server with both FlightOpsService and
// FlightLookupService registered against st.
func NewServer(st *store.Store) *grpc.Server {
	srv := grpc.NewServer()
	opsv1.RegisterFlightOpsServiceServer(srv, NewOpsServer(st))
	lookupv1.RegisterFlightLookupServiceServer(srv, NewLookupServer(st))
	return srv
}
