// Package http implements a generic configurable HTTP adapter.
// When an agent run triggers this adapter, it makes a single HTTP POST
// (or configured method) to the configured URL with a JSON payload.
// This is not a process adapter — BuildCommand returns an error.
package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"abyss-view/pkg/adapter"
)

// Adapter implements adapter.ServerAdapter for generic HTTP webhook invocations.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "http" }

// BuildCommand returns an error — the HTTP adapter is not process-based.
// Use RunHTTP to execute the HTTP call directly.
func (a *Adapter) BuildCommand(_ adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	return nil, fmt.Errorf("http adapter: not a process adapter — invoke via HTTP client")
}

// ParseOutput interprets a single line from an HTTP response body stream.
// Lines that look like JSON are decoded; others are treated as plain text.
func (a *Adapter) ParseOutput(line []byte) (*adapter.OutputEvent, error) {
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" {
		return nil, nil
	}

	// Strip SSE data prefix if present.
	if after, ok := strings.CutPrefix(trimmed, "data: "); ok {
		trimmed = strings.TrimSpace(after)
	}
	if trimmed == "[DONE]" {
		return &adapter.OutputEvent{Type: "done"}, nil
	}

	if strings.HasPrefix(trimmed, "{") {
		var ev map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &ev); err == nil {
			evType, _ := ev["type"].(string)
			if evType == "" {
				evType = "text"
			}
			content, _ := ev["content"].(string)
			return &adapter.OutputEvent{
				Type:     evType,
				Content:  strings.TrimSpace(content),
				Metadata: ev,
			}, nil
		}
	}

	return &adapter.OutputEvent{Type: "text", Content: trimmed}, nil
}

func (a *Adapter) SupportedModels() []adapter.ModelInfo {
	return []adapter.ModelInfo{}
}

// BuildRequestPayload constructs the JSON request body for a generic HTTP
// webhook call. When cfg.CommunicationLanguage is set the "language" field is
// included in the payload so the remote endpoint can enforce response language.
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
