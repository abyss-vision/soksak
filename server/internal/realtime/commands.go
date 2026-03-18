package realtime

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"syscall"
	"time"

	"abyss-view/internal/adapters/process"
)

// HandleClientMessage dispatches an incoming raw WebSocket frame from client.
func HandleClientMessage(client *Client, raw []byte, hub *Hub, pm *process.Manager) error {
	var msg WebSocketMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return sendError(client, "", "parse_error", "invalid JSON")
	}

	switch msg.Type {
	case CmdAgentStdinWrite:
		return handleStdinWrite(client, msg, pm)

	case CmdAgentStdinSignal:
		return handleStdinSignal(client, msg, pm)

	case CmdSubscribe:
		return handleSubscribe(client, msg)

	case CmdUnsubscribe:
		return handleUnsubscribe(client, msg)

	case CmdIssueUpdate:
		// Issue update is handled by the HTTP API; acknowledge receipt only.
		slog.Debug("realtime: received issue.update command (no-op)", "actorID", client.actorID)
		return nil

	default:
		slog.Warn("realtime: unknown command type", "type", msg.Type, "actorID", client.actorID)
		return sendError(client, msg.ID, "unknown_type", fmt.Sprintf("unknown command type: %s", msg.Type))
	}
}

func handleStdinWrite(client *Client, msg WebSocketMessage, pm *process.Manager) error {
	var payload StdinWritePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return sendError(client, msg.ID, "bad_payload", "agent.stdin.write requires runId and data")
	}
	if payload.RunID == "" || payload.Data == "" {
		return sendError(client, msg.ID, "bad_payload", "runId and data are required")
	}
	if err := pm.WriteStdin(payload.RunID, payload.Data); err != nil {
		return sendError(client, msg.ID, "write_failed", err.Error())
	}
	return nil
}

func handleStdinSignal(client *Client, msg WebSocketMessage, pm *process.Manager) error {
	var payload StdinSignalPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return sendError(client, msg.ID, "bad_payload", "agent.stdin.signal requires runId and signal")
	}
	if payload.RunID == "" || payload.Signal == "" {
		return sendError(client, msg.ID, "bad_payload", "runId and signal are required")
	}

	sig, err := parseSignal(payload.Signal)
	if err != nil {
		return sendError(client, msg.ID, "bad_signal", err.Error())
	}
	if err := pm.Signal(payload.RunID, sig); err != nil {
		return sendError(client, msg.ID, "signal_failed", err.Error())
	}
	return nil
}

func handleSubscribe(client *Client, msg WebSocketMessage) error {
	var payload SubscribePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil || payload.Channel == "" {
		return sendError(client, msg.ID, "bad_payload", "subscribe requires a channel")
	}
	client.subs[payload.Channel] = true
	slog.Debug("realtime: client subscribed", "channel", payload.Channel, "actorID", client.actorID)
	return nil
}

func handleUnsubscribe(client *Client, msg WebSocketMessage) error {
	var payload SubscribePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil || payload.Channel == "" {
		return sendError(client, msg.ID, "bad_payload", "unsubscribe requires a channel")
	}
	delete(client.subs, payload.Channel)
	slog.Debug("realtime: client unsubscribed", "channel", payload.Channel, "actorID", client.actorID)
	return nil
}

// sendError enqueues an error response message into the client's send channel.
func sendError(client *Client, msgID, code, detail string) error {
	type errorPayload struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}
	payload, _ := json.Marshal(errorPayload{Code: code, Detail: detail})
	resp := WebSocketMessage{
		Type:      "error",
		ID:        msgID,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
	data, _ := json.Marshal(resp)
	select {
	case client.send <- data:
	default:
	}
	return nil
}

// parseSignal converts a signal name string to an os.Signal.
func parseSignal(name string) (os.Signal, error) {
	switch name {
	case "SIGTERM":
		return syscall.SIGTERM, nil
	case "SIGKILL":
		return syscall.SIGKILL, nil
	case "SIGINT":
		return syscall.SIGINT, nil
	case "SIGHUP":
		return syscall.SIGHUP, nil
	case "SIGUSR1":
		return syscall.SIGUSR1, nil
	case "SIGUSR2":
		return syscall.SIGUSR2, nil
	default:
		return nil, fmt.Errorf("unsupported signal: %s", name)
	}
}
