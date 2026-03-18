package pi

import (
	"encoding/json"
	"strings"

	"abyss-view/pkg/adapter"
)

// Adapter implements adapter.ServerAdapter for the Pi CLI.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "pi_local" }

func (a *Adapter) BuildCommand(cfg adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	if directive := adapter.LanguageDirective(cfg.CommunicationLanguage); directive != "" {
		cfg.Prompt = directive + cfg.Prompt
	}

	command := "pi"
	if override, ok := cfg.ExtraArgs["command"]; ok && override != "" {
		command = override
	}
	args := []string{"run", "--output-format", "jsonl"}
	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}
	for k, v := range cfg.ExtraArgs {
		if k == "command" {
			continue
		}
		args = append(args, k, v)
	}
	return &adapter.CommandSpec{Command: command, Args: args, Env: envSlice(cfg.EnvVars), WorkDir: cfg.WorkDir, Stdin: cfg.Prompt}, nil
}

func (a *Adapter) ParseOutput(line []byte) (*adapter.OutputEvent, error) {
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" {
		return nil, nil
	}
	var ev map[string]interface{}
	if err := json.Unmarshal(line, &ev); err != nil {
		return &adapter.OutputEvent{Type: "text", Content: trimmed}, nil
	}
	evType, _ := ev["type"].(string)
	// Skip internal protocol frames.
	switch evType {
	case "response", "extension_ui_request", "extension_ui_response", "extension_error":
		return nil, nil
	}
	content := ""
	switch evType {
	case "turn_end":
		if msg, ok := ev["message"].(map[string]interface{}); ok {
			content = extractTextContent(msg["content"])
		}
	case "message_update":
		if ae, ok := ev["assistantMessageEvent"].(map[string]interface{}); ok {
			if t, _ := ae["type"].(string); t == "text_delta" {
				content, _ = ae["delta"].(string)
			}
		}
	}
	if evType == "" {
		evType = "text"
	}
	return &adapter.OutputEvent{Type: evType, Content: strings.TrimSpace(content), Metadata: ev}, nil
}

func (a *Adapter) SupportedModels() []adapter.ModelInfo {
	return []adapter.ModelInfo{
		{ID: "pi-3", Name: "Pi 3", Provider: "inflection"},
	}
}

func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
