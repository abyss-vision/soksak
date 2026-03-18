package realtime

import (
	"encoding/json"
	"log/slog"
	"time"

	"nhooyr.io/websocket"
)

// Client represents a single connected WebSocket client.
type Client struct {
	conn      *websocket.Conn
	companyID string
	actorType string
	actorID   string
	subs      map[string]bool
	send      chan []byte
}

// Hub maintains the registry of active clients and manages pub/sub routing.
type Hub struct {
	// clients is the set of all connected clients.
	clients map[*Client]bool

	// rooms maps companyID -> set of clients subscribed to that company.
	rooms map[string]map[*Client]bool

	// register queues a client for registration.
	register chan *Client

	// unregister queues a client for removal.
	unregister chan *Client

	// broadcast sends a message to all clients in a company room.
	broadcast chan companyMessage
}

type companyMessage struct {
	companyID string
	data      []byte
}

// NewHub creates and returns an uninitialised Hub. Call Run() to start it.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan companyMessage, 1024),
	}
}

// Run starts the hub event loop. It blocks until the context is cancelled or
// the channels are closed, so it should be started in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if h.rooms[client.companyID] == nil {
				h.rooms[client.companyID] = make(map[*Client]bool)
			}
			h.rooms[client.companyID][client] = true
			slog.Debug("realtime: client registered",
				"companyID", client.companyID,
				"actorType", client.actorType,
				"actorID", client.actorID,
			)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if room, ok := h.rooms[client.companyID]; ok {
					delete(room, client)
					if len(room) == 0 {
						delete(h.rooms, client.companyID)
					}
				}
				close(client.send)
				slog.Debug("realtime: client unregistered",
					"companyID", client.companyID,
					"actorType", client.actorType,
					"actorID", client.actorID,
				)
			}

		case msg := <-h.broadcast:
			room, ok := h.rooms[msg.companyID]
			if !ok {
				continue
			}
			for client := range room {
				select {
				case client.send <- msg.data:
				default:
					// Slow client — drop the message to avoid blocking the hub.
					slog.Warn("realtime: dropped message for slow client",
						"companyID", client.companyID,
						"actorID", client.actorID,
					)
				}
			}
		}
	}
}

// Register enqueues a client for registration with the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister enqueues a client for removal from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// PublishToCompany marshals msg and broadcasts it to every client in the
// company room. It is safe to call from any goroutine.
func (h *Hub) PublishToCompany(companyID string, msg WebSocketMessage) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("realtime: failed to marshal message", "type", msg.Type, "err", err)
		return
	}
	h.broadcast <- companyMessage{companyID: companyID, data: data}
}
