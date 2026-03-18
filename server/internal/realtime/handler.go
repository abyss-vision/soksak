package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"nhooyr.io/websocket"

	"soksak/internal/adapters/process"
	"soksak/internal/auth"
)

const (
	sendBufSize    = 256
	writeTimeout   = 10 * time.Second
	pongTimeout    = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 32 * 1024 // 32 KiB
)

// WebSocketHandler upgrades HTTP connections to WebSocket and wires up the
// client to the hub and process manager.
func WebSocketHandler(hub *Hub, pm *process.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Resolve actor from context (populated by auth middleware) or fall back
		// to anonymous using query param token for agents.
		actor, _ := auth.ActorFromContext(r.Context())

		companyID := actor.CompanyID
		if companyID == "" {
			// Try to pull companyID from URL path or query when actor has none.
			companyID = r.URL.Query().Get("companyId")
		}
		if companyID == "" {
			http.Error(w, "missing company context", http.StatusUnauthorized)
			return
		}

		// Accept the WebSocket upgrade.
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: false,
			// Allow same-origin and cross-origin from trusted clients.
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			slog.Warn("realtime: websocket accept failed", "err", err)
			return
		}
		conn.SetReadLimit(maxMessageSize)

		client := &Client{
			conn:      conn,
			companyID: companyID,
			actorType: string(actor.Type),
			actorID:   actor.ID,
			subs:      make(map[string]bool),
			send:      make(chan []byte, sendBufSize),
		}

		hub.Register(client)
		defer hub.Unregister(client)

		// Run read/write pumps concurrently; return when either exits.
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			writePump(ctx, client)
		}()

		readPump(ctx, client, hub, pm, cancel)
		<-done
	}
}

// readPump reads messages from the WebSocket connection and dispatches them.
// It returns when the connection closes or the context is cancelled.
func readPump(ctx context.Context, client *Client, hub *Hub, pm *process.Manager, cancel context.CancelFunc) {
	defer cancel()
	for {
		_, data, err := client.conn.Read(ctx)
		if err != nil {
			// Connection closed or context cancelled — normal exit path.
			slog.Debug("realtime: read pump closed",
				"actorID", client.actorID,
				"companyID", client.companyID,
				"err", err,
			)
			return
		}

		if handleErr := HandleClientMessage(client, data, hub, pm); handleErr != nil {
			slog.Warn("realtime: command handler error",
				"actorID", client.actorID,
				"err", handleErr,
			)
		}
	}
}

// writePump drains client.send and forwards messages to the WebSocket connection.
// It also sends periodic pings to detect dead connections.
func writePump(ctx context.Context, client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = client.conn.Close(websocket.StatusNormalClosure, "server shutting down")
			return

		case data, ok := <-client.send:
			if !ok {
				// Hub closed the channel — client was unregistered.
				_ = client.conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := client.conn.Write(writeCtx, websocket.MessageText, data)
			cancel()
			if err != nil {
				slog.Debug("realtime: write pump error", "actorID", client.actorID, "err", err)
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := client.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				slog.Debug("realtime: ping failed", "actorID", client.actorID, "err", err)
				return
			}
		}
	}
}

// HubPublisher wraps Hub to satisfy the process.Publisher interface.
type HubPublisher struct {
	Hub       *Hub
}

func (p *HubPublisher) PublishRunLog(companyID, runID, stream, line string) {
	payload, _ := json.Marshal(RunLogPayload{
		RunID:  runID,
		Stream: stream,
		Line:   line,
	})
	p.Hub.PublishToCompany(companyID, WebSocketMessage{
		Type:      EventHeartbeatRunLog,
		CompanyID: companyID,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	})
}

func (p *HubPublisher) PublishRunCompleted(companyID, runID string, exitCode int) {
	payload, _ := json.Marshal(RunCompletedPayload{
		RunID:    runID,
		ExitCode: exitCode,
	})
	p.Hub.PublishToCompany(companyID, WebSocketMessage{
		Type:      EventRunCompleted,
		CompanyID: companyID,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	})
}
