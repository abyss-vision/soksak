package gemini

import (
	"encoding/json"
	"strings"

	"abyss-view/pkg/adapter"
)

// Adapter implements adapter.ServerAdapter for the Gemini CLI.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "gemini_local" }

func (a *Adapter) BuildCommand(cfg adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	if directive := adapter.LanguageDirective(cfg.CommunicationLanguage); directive != "" {
		cfg.Prompt = directive + cfg.Prompt
	}

	command := "gemini"
	if override, ok := cfg.ExtraArgs["command"]; ok && override != "" {
		command = override
	}
	args := []string{"--output-format", "stream-json", "--yolo"}
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
	case "assistant":
		texts := collectMessageText(ev["message"])
		content = strings.Join(texts, "")
	case "result":
		content = anyString(firstOf(ev, "result", "text", "response"))
	case "error":
		content = extractErrorText(firstOf(ev, "error", "message", "detail"))
		evType = "error"
	}
	if evType == "" {
		evType = "text"
	}
	return &adapter.OutputEvent{Type: evType, Content: strings.TrimSpace(content), Metadata: ev}, nil
}

func (a *Adapter) SupportedModels() []adapter.ModelInfo {
	return []adapter.ModelInfo{
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: "google"},
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", Provider: "google"},
		{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Provider: "google"},
	}
}

func envSlice(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
