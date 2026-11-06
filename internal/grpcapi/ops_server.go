package grpcapi

import (
	"context"
	"io"
	"time"

	"skygate/internal/domain"
	opsv1 "skygate/internal/grpcapi/ops/v1"
	"skygate/internal/store"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OpsServer implements opsv1.FlightOpsServiceServer against the given Store.
type OpsServer struct {
	opsv1.UnimplementedFlightOpsServiceServer
	store          *store.Store
	streamInterval time.Duration
}

// NewOpsServer returns an OpsServer backed by st, polling it every 2 seconds
// for StreamFleetStatus.
func NewOpsServer(st *store.Store) *OpsServer {
	return &OpsServer{store: st, streamInterval: 2 * time.Second}
}

func toProtoFlight(f domain.Flight) *opsv1.Flight {
	return &opsv1.Flight{
		Id:              f.ID,
		FlightNumber:    f.FlightNumber,
		OriginCode:      f.OriginCode,
		DestinationCode: f.DestinationCode,
		AircraftId:      f.AircraftID,
		Status:          string(f.Status),
		Gate:            f.Gate,
	}
}

// AssignGate sets the flight's gate and persists it, or returns NotFound if
// the flight doesn't exist.
func (s *OpsServer) AssignGate(ctx context.Context, req *opsv1.AssignGateRequest) (*opsv1.Flight, error) {
	f, ok := s.store.GetFlight(req.GetFlightId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "flight %q not found", req.GetFlightId())
	}
	f.Gate = req.GetGate()
	s.store.PutFlight(f)
	return toProtoFlight(f), nil
}

// UpdateFlightStatus transitions the flight to a new status and persists it.
// Returns NotFound if the flight doesn't exist, InvalidArgument if the
// status isn't one of domain.ValidFlightStatuses. Mirrors the REST PATCH
// handler's validation and its rule that Position is only meaningful while
// airborne or holding, so the two write paths can't leave the store in
// inconsistent states.
func (s *OpsServer) UpdateFlightStatus(ctx context.Context, req *opsv1.UpdateFlightStatusRequest) (*opsv1.Flight, error) {
	f, ok := s.store.GetFlight(req.GetFlightId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "flight %q not found", req.GetFlightId())
	}
	if !domain.IsValidFlightStatus(req.GetStatus()) {
		return nil, status.Errorf(codes.InvalidArgument, "status %q is not a known flight status", req.GetStatus())
	}
	f.Status = domain.FlightStatus(req.GetStatus())
	if f.Status != domain.StatusAirborne && f.Status != domain.StatusHolding {
		f.Position = nil
	}
	s.store.PutFlight(f)
	return toProtoFlight(f), nil
}

// StreamFleetStatus sends every flight in the store on each tick of
// s.streamInterval until the client disconnects or the stream errors.
func (s *OpsServer) StreamFleetStatus(_ *opsv1.Empty, stream opsv1.FlightOpsService_StreamFleetStatusServer) error {
	ticker := time.NewTicker(s.streamInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			for _, f := range s.store.ListFlights() {
				if err := stream.Send(toProtoFlight(f)); err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
			}
		}
	}
}
