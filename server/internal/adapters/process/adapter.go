package process

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"abyss-view/pkg/adapter"
)

// ProcessAdapter implements adapter.ServerAdapter for generic command-based
// agents. It delegates process lifecycle management to the Manager.
type ProcessAdapter struct {
	manager *Manager
}

// NewProcessAdapter creates a ProcessAdapter backed by the given Manager.
func NewProcessAdapter(m *Manager) *ProcessAdapter {
	return &ProcessAdapter{manager: m}
}

// Name returns the adapter type identifier.
func (a *ProcessAdapter) Name() string {
	return "process"
}

// BuildCommand translates an AdapterConfig into a CommandSpec.
// The config must include the command in EnvVars["_command"] or the AdapterType
// is used as a hint. For the generic process adapter the command comes from
// AdapterConfig.AdapterType which is expected to be an executable name, or
// callers may override via ExtraArgs["command"].
func (a *ProcessAdapter) BuildCommand(cfg adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	command := cfg.AdapterType
	// Allow an explicit command override via ExtraArgs.
	if override, ok := cfg.ExtraArgs["command"]; ok && override != "" {
		command = override
	}
	if command == "" {
		return nil, fmt.Errorf("process adapter: command is required")
	}

	// Build args list from ExtraArgs (excluding the "command" key).
	var args []string
	for k, v := range cfg.ExtraArgs {
		if k == "command" {
			continue
		}
		args = append(args, k, v)
	}

	// Convert EnvVars map to KEY=VALUE slice.
	env := make([]string, 0, len(cfg.EnvVars))
	for k, v := range cfg.EnvVars {
		env = append(env, k+"="+v)
	}

	return &adapter.CommandSpec{
		Command: command,
		Args:    args,
		Env:     env,
		WorkDir: cfg.WorkDir,
	}, nil
}

// ParseOutput interprets a single line of subprocess stdout output.
// Lines that look like JSON are decoded and classified by their "type" field.
// Plain-text lines are returned as "text" events.
func (a *ProcessAdapter) ParseOutput(line []byte) (*adapter.OutputEvent, error) {
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" {
		return nil, nil
	}

	if trimmed[0] == '{' {
		var raw map[string]interface{}
		if err := json.Unmarshal(line, &raw); err == nil {
			eventType := "text"
			if t, ok := raw["type"].(string); ok && t != "" {
				eventType = t
			}
			content := ""
			if c, ok := raw["content"].(string); ok {
				content = c
			}
			return &adapter.OutputEvent{
				Type:     eventType,
				Content:  content,
				Metadata: raw,
			}, nil
		}
	}

	// Fall back to plain-text event.
	return &adapter.OutputEvent{
		Type:    "text",
		Content: trimmed,
	}, nil
}

// SupportedModels returns an empty slice — the generic process adapter does not
// advertise specific models.
func (a *ProcessAdapter) SupportedModels() []adapter.ModelInfo {
	return []adapter.ModelInfo{}
}

// Run is a convenience helper that combines BuildCommand and Manager.Start.
// It returns the assigned runID for the spawned subprocess.
func (a *ProcessAdapter) Run(ctx context.Context, cfg adapter.AdapterConfig, agentID, companyID string, pub Publisher) (string, error) {
	spec, err := a.BuildCommand(cfg)
	if err != nil {
		return "", err
	}

	return a.manager.Start(ctx, ProcessConfig{
		Command:   spec.Command,
		Args:      spec.Args,
		Env:       spec.Env,
		WorkDir:   spec.WorkDir,
		AgentID:   agentID,
		CompanyID: companyID,
	}, pub)
}
