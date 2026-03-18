package cursor

import (
	"encoding/json"
	"strings"

	"abyss-view/pkg/adapter"
)

// Adapter implements adapter.ServerAdapter for the Cursor background agent CLI.
type Adapter struct{}

// New returns a new Cursor adapter.
func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "cursor" }

func (a *Adapter) BuildCommand(cfg adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	if directive := adapter.LanguageDirective(cfg.CommunicationLanguage); directive != "" {
		cfg.Prompt = directive + cfg.Prompt
	}

	command := "cursor"
	if override, ok := cfg.ExtraArgs["command"]; ok && override != "" {
		command = override
	}
	args := []string{"--background-agent", "--output-format", "stream-json"}
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
	trimmed := strings.TrimSpace(normalizeLine(string(line)))
	if trimmed == "" {
		return nil, nil
	}
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &ev); err != nil {
		return &adapter.OutputEvent{Type: "text", Content: trimmed}, nil
	}
	evType, _ := ev["type"].(string)
	content := ""
	switch evType {
	case "assistant":
		texts := collectMessageText(ev["message"])
		content = strings.Join(texts, "")
	case "result":
		content, _ = ev["result"].(string)
	case "error":
		content = extractErrorText(firstOf(ev, "message", "error", "detail"))
		evType = "error"
	}
	if evType == "" {
		evType = "text"
	}
	return &adapter.OutputEvent{Type: evType, Content: strings.TrimSpace(content), Metadata: ev}, nil
}

func (a *Adapter) SupportedModels() []adapter.ModelInfo {
	return []adapter.ModelInfo{
		{ID: "claude-3-7-sonnet", Name: "Claude 3.7 Sonnet", Provider: "anthropic"},
		{ID: "gpt-4o", Name: "GPT-4o", Provider: "openai"},
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: "google"},
	}
}

func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
