package plugins

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
)

// NodeMessage is the envelope exchanged over stdin/stdout with a Node.js plugin.
type NodeMessage struct {
	// Type identifies the message kind (e.g. "event", "result", "error", "info").
	Type string `json:"type"`
	// Payload carries the message-specific data.
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NodeEventPayload is the payload for "event" messages sent to the Node.js process.
type NodeEventPayload struct {
	EventType   string                 `json:"eventType"`
	CompanyUUID string                 `json:"companyUuid"`
	Data        map[string]interface{} `json:"data"`
}

// NodeResultPayload is the payload returned by the Node.js process after handling an event.
type NodeResultPayload struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// NodeBridge wraps a Node.js child process and communicates with it via
// newline-delimited JSON over stdin/stdout.
type NodeBridge struct {
	scriptPath string

	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	stderr *bufio.Scanner
}

// NewNodeBridge creates a NodeBridge for the given .js plugin script.
// Call Start() to spawn the process.
func NewNodeBridge(scriptPath string) *NodeBridge {
	return &NodeBridge{scriptPath: scriptPath}
}

// Start spawns the Node.js process. Returns an error if the process cannot start.
func (b *NodeBridge) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cmd != nil {
		return fmt.Errorf("nodebridge: already started")
	}

	cmd := exec.CommandContext(ctx, "node", b.scriptPath) //nolint:gosec
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("nodebridge: stdin pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("nodebridge: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("nodebridge: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("nodebridge: start node: %w", err)
	}

	b.cmd = cmd
	b.stdin = stdinPipe
	b.stdout = bufio.NewScanner(stdoutPipe)
	b.stderr = bufio.NewScanner(stderrPipe)

	// Forward stderr to structured logging.
	go func() {
		for b.stderr.Scan() {
			slog.Warn("nodebridge stderr", "script", b.scriptPath, "line", b.stderr.Text())
		}
	}()

	return nil
}

// Stop sends a {"type":"shutdown"} message to the Node.js process, then closes
// stdin to signal EOF, and waits for the process to exit.
func (b *NodeBridge) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cmd == nil {
		return nil
	}

	// Best-effort shutdown signal.
	_ = b.sendLocked(NodeMessage{Type: "shutdown"})
	_ = b.stdin.Close()

	err := b.cmd.Wait()
	b.cmd = nil
	return err
}

// SendEvent sends a domain event to the Node.js plugin and waits for a result.
// The call blocks until the plugin responds or ctx is cancelled.
func (b *NodeBridge) SendEvent(ctx context.Context, eventType, companyUUID string, data map[string]interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cmd == nil {
		return fmt.Errorf("nodebridge: not started")
	}

	payload, err := json.Marshal(NodeEventPayload{
		EventType:   eventType,
		CompanyUUID: companyUUID,
		Data:        data,
	})
	if err != nil {
		return fmt.Errorf("nodebridge: marshal event: %w", err)
	}

	msg := NodeMessage{Type: "event", Payload: payload}
	if err := b.sendLocked(msg); err != nil {
		return err
	}

	// Read one response line.
	type readResult struct {
		line string
		err  error
	}
	ch := make(chan readResult, 1)
	go func() {
		if b.stdout.Scan() {
			ch <- readResult{line: b.stdout.Text()}
		} else {
			ch <- readResult{err: fmt.Errorf("nodebridge: stdout closed")}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return res.err
		}
		return b.parseResult(res.line)
	}
}

// sendLocked writes a JSON message to the Node.js process. Must be called
// with b.mu held.
func (b *NodeBridge) sendLocked(msg NodeMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("nodebridge: marshal message: %w", err)
	}
	data = append(data, '\n')
	if _, err := b.stdin.Write(data); err != nil {
		return fmt.Errorf("nodebridge: write to stdin: %w", err)
	}
	return nil
}

// parseResult parses a single response line from the Node.js process.
func (b *NodeBridge) parseResult(line string) error {
	var msg NodeMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return fmt.Errorf("nodebridge: parse response %q: %w", line, err)
	}

	if msg.Type == "error" {
		var errPayload struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(msg.Payload, &errPayload); err == nil && errPayload.Message != "" {
			return fmt.Errorf("nodebridge: plugin error: %s", errPayload.Message)
		}
		return fmt.Errorf("nodebridge: plugin returned error")
	}

	if msg.Type == "result" {
		var result NodeResultPayload
		if err := json.Unmarshal(msg.Payload, &result); err != nil {
			return fmt.Errorf("nodebridge: parse result payload: %w", err)
		}
		if !result.OK {
			return fmt.Errorf("nodebridge: plugin handler failed: %s", result.Error)
		}
		return nil
	}

	// Ignore unknown message types (e.g. "log").
	return nil
}
