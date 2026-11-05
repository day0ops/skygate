package ws

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"skygate/internal/domain"
	"skygate/internal/store"
)

func TestHandlerSendsSnapshot(t *testing.T) {
	st := store.New()
	st.PutFlight(domain.Flight{ID: "flt_1", FlightNumber: "SK100", Status: domain.StatusAirborne})
	hub := NewHub()

	srv := httptest.NewServer(Handler(st, hub))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(msg), "SK100") {
		t.Fatalf("expected snapshot to contain flight SK100, got: %s", msg)
	}
}

func TestHubBroadcast(t *testing.T) {
	st := store.New()
	hub := NewHub()

	srv := httptest.NewServer(Handler(st, hub))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond) // let the server register the client
	hub.Broadcast(domain.Flight{ID: "flt_2", FlightNumber: "SK200", Status: domain.StatusBoarding})

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(msg), "SK200") {
		t.Fatalf("expected broadcast to contain flight SK200, got: %s", msg)
	}
}

// TestConcurrentConnectAndBroadcastDoesNotRace drives many concurrent
// connections (each sending a multi-message snapshot) against a goroutine
// hammering Broadcast, with no synchronization sleep. gorilla/websocket
// allows only one concurrent writer per connection; run with -race, this
// catches a regression where a client's snapshot write overlaps a Broadcast
// write to that same connection.
func TestConcurrentConnectAndBroadcastDoesNotRace(t *testing.T) {
	st := store.New()
	for i := 0; i < 20; i++ {
		st.PutFlight(domain.Flight{ID: domain.NewID("flt"), FlightNumber: "SK1", Status: domain.StatusAirborne})
	}
	hub := NewHub()

	srv := httptest.NewServer(Handler(st, hub))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				hub.Broadcast(domain.Flight{ID: "flt_bcast", FlightNumber: "SK900", Status: domain.StatusBoarding})
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("dial: %v", err)
				return
			}
			defer conn.Close()
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if _, _, err := conn.ReadMessage(); err != nil {
				t.Errorf("read: %v", err)
			}
		}()
	}
	wg.Wait()
	close(stop)
}
