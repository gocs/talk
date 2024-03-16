package talk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

type Message struct {
	TransactionID string `json:"tid"`
	DestinationID string `json:"dst"`
	SourceID      string `json:"src"`
	Status        uint32 `json:"status"`
	Value         string `json:"val"`
}

func (h *Hub) Run(ctx context.Context) error {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			m := &Message{}
			if err := json.Unmarshal(message, m); err != nil {
				slog.Error("Run", "err", err)
				continue
			}

			for client := range h.clients {
				if m.DestinationID == client.user {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func JSONErr(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	r := map[string]string{"error": error, "code": fmt.Sprintf("%d", code)}
	if err := json.NewEncoder(w).Encode(r); err != nil {
		slog.Error("JSONRes", "err", err)
	}
}

func JSONRes(w http.ResponseWriter, a any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(a); err != nil {
		slog.Error("JSONRes", "err", err)
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	if user == "" {
		JSONErr(w, "path value: invalid user", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		JSONErr(w, "upgrade: cannot upgrade", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	client := NewClient(h, conn, user)
	h.register <- client
	defer func() { h.unregister <- client }()

	go client.WritePump()
	client.ReadPump()
}
