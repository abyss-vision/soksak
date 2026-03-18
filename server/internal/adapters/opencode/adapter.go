package opencode

import (
	"encoding/json"
	"strings"

	"soksak/pkg/adapter"
)

// Adapter implements adapter.ServerAdapter for the opencode CLI.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "opencode_local" }

func (a *Adapter) BuildCommand(cfg adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	if directive := adapter.LanguageDirective(cfg.CommunicationLanguage); directive != "" {
		cfg.Prompt = directive + cfg.Prompt
	}

	command := "opencode"
	if override, ok := cfg.ExtraArgs["command"]; ok && override != "" {
		command = override
	}
	args := []string{"run", "--output-format", "stream-json"}
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
	content := ""
	switch evType {
	case "text":
		if part, ok := ev["part"].(map[string]interface{}); ok {
			content, _ = part["text"].(string)
		}
	case "error":
		content = errorText(firstOf(ev, "error", "message"))
		evType = "error"
	}
	if evType == "" {
		evType = "text"
	}
	return &adapter.OutputEvent{Type: evType, Content: strings.TrimSpace(content), Metadata: ev}, nil
}

func (a *Adapter) SupportedModels() []adapter.ModelInfo {
	// OpenCode supports dynamic model discovery; static list is intentionally minimal.
	return []adapter.ModelInfo{
		{ID: "anthropic/claude-sonnet-4-5", Name: "Claude Sonnet 4.5", Provider: "anthropic"},
		{ID: "openai/gpt-4o", Name: "GPT-4o", Provider: "openai"},
	}
}

func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
