package grpcapi

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"skygate/internal/domain"
	lookupv1 "skygate/internal/grpcapi/lookup/v1"
	"skygate/internal/store"
)

func TestServerServesLookupService(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", Status: domain.StatusAirborne, DelayMinutes: 5})

	lis := bufconn.Listen(1024 * 1024)
	srv := NewServer(st)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := lookupv1.NewFlightLookupServiceClient(conn)
	resp, err := client.GetFlightStatus(context.Background(), &lookupv1.GetFlightStatusRequest{FlightId: "flt_1"})
	if err != nil {
		t.Fatalf("GetFlightStatus: %v", err)
	}
	if resp.Status != "airborne" || resp.DelayMinutes != 5 {
		t.Fatalf("got %+v, want status=airborne delayMinutes=5", resp)
	}
}
