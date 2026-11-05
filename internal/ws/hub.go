package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"skygate/internal/domain"
	"skygate/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// writeTimeout bounds every write to a client connection. Broadcast runs
// synchronously inside the simulation engine's Tick, holding engine.mu for
// its duration - without a deadline, one stalled client (backgrounded tab,
// sleeping laptop, flaky Wi-Fi) would block WriteMessage indefinitely and
// freeze the simulation and every other client's feed along with it.
const writeTimeout = 5 * time.Second

type flightUpdateMessage struct {
	Type   string        `json:"type"`
	Flight domain.Flight `json:"flight"`
}

// Hub fans out flight-update events to every connected WebSocket client.
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

// NewHub returns a new Hub ready to manage WebSocket clients.
func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

// Broadcast sends f to every currently connected client. Safe to call from
// the simulation engine's goroutine.
func (h *Hub) Broadcast(f domain.Flight) {
	msg, err := json.Marshal(flightUpdateMessage{Type: "flight_update", Flight: f})
	if err != nil {
		log.Printf("ws: marshal update: %v", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		_ = c.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			c.Close()
			delete(h.clients, c)
		}
	}
}

// register adds c to the client set and writes every message in snapshot to
// it, all under h.mu - the same lock Broadcast holds for its own writes.
// gorilla/websocket allows at most one concurrent writer per connection, so
// a newly-connecting client's snapshot must never overlap with a Broadcast
// write to that same connection; sharing the lock is what guarantees that.
func (h *Hub) register(c *websocket.Conn, snapshot [][]byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
	for _, msg := range snapshot {
		_ = c.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(h.clients, c)
			return err
		}
	}
	return nil
}

func (h *Hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
}

// Handler upgrades the connection, sends the full current flight snapshot,
// then blocks until the client disconnects. Incremental updates arrive via
// Broadcast, called from elsewhere (the simulation engine's subscriber).
func Handler(st *store.Store, h *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws: upgrade: %v", err)
			return
		}
		defer conn.Close()

		flights := st.ListFlights()
		snapshot := make([][]byte, 0, len(flights))
		for _, f := range flights {
			msg, err := json.Marshal(flightUpdateMessage{Type: "flight_update", Flight: f})
			if err != nil {
				log.Printf("ws: marshal snapshot: %v", err)
				continue
			}
			snapshot = append(snapshot, msg)
		}
		if err := h.register(conn, snapshot); err != nil {
			return
		}
		defer h.remove(conn)

		// Drain and discard client messages; this is also how we detect
		// disconnects (ReadMessage errors once the client closes).
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}
}
