package adapters

import (
	"soksak/internal/adapters/claude"
	"soksak/internal/adapters/codex"
	"soksak/internal/adapters/cursor"
	"soksak/internal/adapters/gemini"
	httpadapter "soksak/internal/adapters/http"
	"soksak/internal/adapters/openclaw"
	"soksak/internal/adapters/opencode"
	"soksak/internal/adapters/pi"
	processadapter "soksak/internal/adapters/process"
	"soksak/pkg/adapter"
)

// RegisterAll registers all built-in AI agent adapters into the provided registry.
// procManager must be non-nil; it backs the generic process adapter.
func RegisterAll(registry *adapter.Registry, procManager *processadapter.Manager) {
	registry.Register(claude.New())
	registry.Register(codex.New())
	registry.Register(cursor.New())
	registry.Register(gemini.New())
	registry.Register(opencode.New())
	registry.Register(pi.New())
	registry.Register(openclaw.New())
	registry.Register(httpadapter.New())
	registry.Register(processadapter.NewProcessAdapter(procManager))
}
