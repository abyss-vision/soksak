package adapters

import (
	"abyss-view/internal/adapters/claude"
	"abyss-view/internal/adapters/codex"
	"abyss-view/internal/adapters/cursor"
	"abyss-view/internal/adapters/gemini"
	httpadapter "abyss-view/internal/adapters/http"
	"abyss-view/internal/adapters/openclaw"
	"abyss-view/internal/adapters/opencode"
	"abyss-view/internal/adapters/pi"
	processadapter "abyss-view/internal/adapters/process"
	"abyss-view/pkg/adapter"
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
