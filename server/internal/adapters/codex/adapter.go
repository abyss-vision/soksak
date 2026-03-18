package codex

import (
	"encoding/json"
	"strings"

	"soksak/pkg/adapter"
)

// Adapter implements adapter.ServerAdapter for the OpenAI Codex CLI.
type Adapter struct{}

// New returns a new Codex adapter.
func New() *Adapter { return &Adapter{} }

// Name returns the adapter type identifier.
func (a *Adapter) Name() string { return "codex_local" }

// BuildCommand constructs the codex CLI invocation.
// codex --full-auto --quiet [--model MODEL] [--approval-policy auto] [extra-args...]
func (a *Adapter) BuildCommand(cfg adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	if directive := adapter.LanguageDirective(cfg.CommunicationLanguage); directive != "" {
		cfg.Prompt = directive + cfg.Prompt
	}

	command := "codex"
	if override, ok := cfg.ExtraArgs["command"]; ok && override != "" {
		command = override
	}

	args := []string{"--full-auto", "--quiet"}

	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}
	args = append(args, "--approval-policy", "auto")

	for k, v := range cfg.ExtraArgs {
		if k == "command" {
			continue
		}
		args = append(args, k, v)
	}

	return &adapter.CommandSpec{
		Command: command,
		Args:    args,
		Env:     envSlice(cfg.EnvVars),
		WorkDir: cfg.WorkDir,
		Stdin:   cfg.Prompt,
	}, nil
}

// ParseOutput interprets a single JSONL line from codex stdout.
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
	case "item.completed":
		if item, ok := ev["item"].(map[string]interface{}); ok {
			if t, _ := item["type"].(string); t == "agent_message" {
				content, _ = item["text"].(string)
			}
		}
	case "error", "turn.failed":
		evType = "error"
		if msg, _ := ev["message"].(string); msg != "" {
			content = msg
		}
	}

	if evType == "" {
		evType = "text"
	}

	return &adapter.OutputEvent{
		Type:     evType,
		Content:  strings.TrimSpace(content),
		Metadata: ev,
	}, nil
}

// SupportedModels returns known Codex / OpenAI models.
func (a *Adapter) SupportedModels() []adapter.ModelInfo {
	return []adapter.ModelInfo{
		{ID: "codex-mini-latest", Name: "Codex Mini", Provider: "openai"},
		{ID: "o4-mini", Name: "o4-mini", Provider: "openai"},
		{ID: "o3", Name: "o3", Provider: "openai"},
		{ID: "o3-mini", Name: "o3-mini", Provider: "openai"},
	}
}

func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
