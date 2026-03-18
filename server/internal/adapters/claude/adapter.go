package claude

import (
	"encoding/json"
	"strings"

	"soksak/pkg/adapter"
)

// Adapter implements adapter.ServerAdapter for Claude Code CLI.
// It invokes: claude --print - --output-format stream-json --verbose [model] [extra-args]
type Adapter struct{}

// New returns a new Claude adapter.
func New() *Adapter { return &Adapter{} }

// Name returns the adapter type identifier.
func (a *Adapter) Name() string { return "claude_local" }

// BuildCommand constructs the claude CLI invocation from cfg.
func (a *Adapter) BuildCommand(cfg adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	if directive := adapter.LanguageDirective(cfg.CommunicationLanguage); directive != "" {
		cfg.Prompt = directive + cfg.Prompt
	}

	command := "claude"
	if override, ok := cfg.ExtraArgs["command"]; ok && override != "" {
		command = override
	}

	args := []string{"--print", "-", "--output-format", "stream-json", "--verbose"}

	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}

	// Optional flags passed through ExtraArgs (skip "command" which is the binary override).
	for k, v := range cfg.ExtraArgs {
		switch k {
		case "command":
			// already consumed
		case "--resume":
			args = append(args, "--resume", v)
		case "--max-turns":
			args = append(args, "--max-turns", v)
		case "--dangerously-skip-permissions":
			if v == "true" {
				args = append(args, "--dangerously-skip-permissions")
			}
		default:
			args = append(args, k, v)
		}
	}

	env := envSlice(cfg.EnvVars)

	return &adapter.CommandSpec{
		Command: command,
		Args:    args,
		Env:     env,
		WorkDir: cfg.WorkDir,
		Stdin:   cfg.Prompt,
	}, nil
}

// ParseOutput interprets a single stream-json line from Claude stdout.
func (a *Adapter) ParseOutput(line []byte) (*adapter.OutputEvent, error) {
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" {
		return nil, nil
	}

	var ev map[string]interface{}
	if err := json.Unmarshal(line, &ev); err != nil {
		// plain text fallback
		return &adapter.OutputEvent{Type: "text", Content: trimmed}, nil
	}

	evType, _ := ev["type"].(string)
	if evType == "" {
		evType = "text"
	}

	content := ""
	switch evType {
	case "assistant":
		if msg, ok := ev["message"].(map[string]interface{}); ok {
			if blocks, ok := msg["content"].([]interface{}); ok {
				var parts []string
				for _, raw := range blocks {
					if b, ok := raw.(map[string]interface{}); ok {
						if bt, _ := b["type"].(string); bt == "text" {
							if t, _ := b["text"].(string); t != "" {
								parts = append(parts, t)
							}
						}
					}
				}
				content = strings.Join(parts, "")
			}
		}
	case "result":
		content, _ = ev["result"].(string)
	}

	return &adapter.OutputEvent{
		Type:     evType,
		Content:  content,
		Metadata: ev,
	}, nil
}

// SupportedModels returns a static list of known Claude models.
func (a *Adapter) SupportedModels() []adapter.ModelInfo {
	return []adapter.ModelInfo{
		{ID: "claude-opus-4-5", Name: "Claude Opus 4.5", Provider: "anthropic"},
		{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", Provider: "anthropic"},
		{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5", Provider: "anthropic"},
		{ID: "claude-opus-4-0", Name: "Claude Opus 4", Provider: "anthropic"},
		{ID: "claude-sonnet-4-0", Name: "Claude Sonnet 4", Provider: "anthropic"},
	}
}

func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
