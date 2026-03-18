// Package openclaw implements the ServerAdapter interface for the OpenClaw gateway.
// Unlike the process-based adapters, OpenClaw communicates over a WebSocket/HTTP
// gateway and does NOT spawn a local subprocess. BuildCommand returns a sentinel
// spec so the process manager is never invoked; callers should use RunHTTP instead.
package openclaw

import (
	"encoding/json"
	"fmt"
	"strings"

	"soksak/pkg/adapter"
)

// Adapter implements adapter.ServerAdapter for the OpenClaw gateway.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "openclaw_gateway" }

// BuildCommand returns an error — OpenClaw is not a process adapter.
// Use the gateway HTTP/WebSocket client directly.
func (a *Adapter) BuildCommand(_ adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	return nil, fmt.Errorf("openclaw_gateway: not a process adapter — use gateway HTTP client")
}

// ParseOutput interprets a single SSE line from the OpenClaw gateway event stream.
// Lines beginning with "data: " are unwrapped per the SSE spec.
func (a *Adapter) ParseOutput(line []byte) (*adapter.OutputEvent, error) {
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" || trimmed == ":" {
		return nil, nil
	}

	// Strip SSE data prefix.
	data := trimmed
	if after, ok := strings.CutPrefix(trimmed, "data: "); ok {
		data = strings.TrimSpace(after)
	}
	if data == "[DONE]" {
		return &adapter.OutputEvent{Type: "done", Content: ""}, nil
	}

	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		return &adapter.OutputEvent{Type: "text", Content: data}, nil
	}

	evType, _ := ev["type"].(string)
	content := ""

	switch evType {
	case "message", "text":
		content, _ = ev["content"].(string)
	case "error":
		content = extractErrText(ev)
		evType = "error"
	case "done", "":
		evType = "done"
	}

	return &adapter.OutputEvent{
		Type:     evType,
		Content:  strings.TrimSpace(content),
		Metadata: ev,
	}, nil
}

func (a *Adapter) SupportedModels() []adapter.ModelInfo {
	// Models are provisioned server-side on the OpenClaw gateway.
	return []adapter.ModelInfo{}
}

// BuildRequestPayload constructs the JSON request body for an OpenClaw gateway
// call. When cfg.CommunicationLanguage is set the "language" field is included
// so the gateway can enforce the response language on its side.
func (a *Adapter) BuildRequestPayload(cfg adapter.AdapterConfig) map[string]interface{} {
	body := map[string]interface{}{
		"prompt":  cfg.Prompt,
		"model":   cfg.Model,
		"workDir": cfg.WorkDir,
	}
	if cfg.CommunicationLanguage != "" {
		body["language"] = cfg.CommunicationLanguage
	}
	if len(cfg.ExtraArgs) > 0 {
		body["extraArgs"] = cfg.ExtraArgs
	}
	return body
}

func extractErrText(ev map[string]interface{}) string {
	for _, key := range []string{"message", "error", "detail"} {
		if s, _ := ev[key].(string); s != "" {
			return strings.TrimSpace(s)
		}
	}
	b, _ := json.Marshal(ev)
	return string(b)
}
