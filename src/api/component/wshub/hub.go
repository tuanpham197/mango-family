// Package wshub — WebSocket hub theo household (research D8):
// client kết nối /ws sau khi authenticate; hub broadcast event invalidation
// {"type":"categories_changed"|...} tới mọi client cùng hộ.
package wshub

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
	// Same-origin: dev qua Vite proxy, prod Go serve web/dist — không cần CORS check phức tạp.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*client]struct{} // householdID → clients
}

func NewHub() *Hub {
	return &Hub{clients: map[uuid.UUID]map[*client]struct{}{}}
}

type Event struct {
	Type string `json:"type"`
}

// Broadcast gửi event tới mọi client của một hộ.
func (h *Hub) Broadcast(householdID uuid.UUID, eventType string) {
	payload, _ := json.Marshal(Event{Type: eventType})
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[householdID] {
		select {
		case c.send <- payload:
		default: // client nghẽn thì bỏ event — FE fallback refetch khi focus
		}
	}
}

func (h *Hub) add(householdID uuid.UUID, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[householdID] == nil {
		h.clients[householdID] = map[*client]struct{}{}
	}
	h.clients[householdID][c] = struct{}{}
}

func (h *Hub) remove(householdID uuid.UUID, c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[householdID], c)
	close(c.send)
}

// ServeWS nâng cấp HTTP → WebSocket cho một client đã authenticate.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, householdID uuid.UUID) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c := &client{conn: conn, send: make(chan []byte, 16)}
	h.add(householdID, c)

	go func() { // write pump
		ticker := time.NewTicker(pingPeriod)
		defer func() { ticker.Stop(); conn.Close() }()
		for {
			select {
			case msg, ok := <-c.send:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	go func() { // read pump (chỉ để nhận pong/close)
		defer h.remove(householdID, c)
		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
	return nil
}
