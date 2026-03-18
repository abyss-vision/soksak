package adapter

// ServerAdapter is the contract every adapter implementation must satisfy.
type ServerAdapter interface {
	// Name returns the unique adapter type identifier (e.g. "claude_local", "process").
	Name() string

	// BuildCommand translates an AdapterConfig into the command-line spec
	// that the process manager will execute.
	BuildCommand(config AdapterConfig) (*CommandSpec, error)

	// ParseOutput interprets a single line of stdout from the spawned process
	// and returns a structured event, or nil if the line should be skipped.
	ParseOutput(line []byte) (*OutputEvent, error)

	// SupportedModels returns the static list of models advertised by this
	// adapter. An empty slice is valid (e.g. adapter uses dynamic model discovery).
	SupportedModels() []ModelInfo
}

// AdapterConfig carries all runtime parameters needed to build a command.
type AdapterConfig struct {
	// AdapterType identifies which adapter handles this run (e.g. "claude_local").
	AdapterType string

	// Model is the LLM model identifier requested for this run.
	Model string

	// WorkDir is the working directory in which the subprocess runs.
	WorkDir string

	// Prompt is the user-facing input text piped to the process via stdin.
	Prompt string

	// ExtraArgs are appended to the command-line after the adapter's own args.
	ExtraArgs map[string]string

	// EnvVars are additional KEY=VALUE environment variables injected into the
	// subprocess environment on top of the inheritable process environment.
	EnvVars map[string]string

	// CommunicationLanguage is the ISO 639-1 language code the AI must respond in.
	// When set, adapters inject a language directive into the prompt or request body.
	// Supported values: "en" (English), "ko" (Korean), "ja" (Japanese).
	// Empty string means no language constraint.
	CommunicationLanguage string
}

// CommandSpec is the fully-resolved command that the process manager launches.
type CommandSpec struct {
	Command string
	Args    []string
	// Env contains KEY=VALUE strings appended to os.Environ().
	Env     []string
	WorkDir string
	// Stdin is optional initial content written to the process stdin immediately
	// after startup. Used by CLI adapters that read the prompt from stdin
	// (e.g. claude --print -). Empty string means no pre-seeded stdin.
	Stdin string
}

// OutputEvent is the structured representation of a single stdout line emitted
// by the subprocess.
type OutputEvent struct {
	// Type classifies the event. Known values: "text", "tool_use",
	// "tool_result", "error", "done".
	Type string

	// Content is the human-readable payload of the event.
	Content string

	// Metadata carries adapter-specific structured data.
	Metadata map[string]interface{}
}

// ModelInfo describes a single model offered by an adapter.
type ModelInfo struct {
	ID       string
	Name     string
	Provider string
}

// languageNames maps ISO 639-1 codes to their full language names used in
// the communication language directive injected into prompts.
var languageNames = map[string]string{
	"en": "English",
	"ko": "Korean (한국어)",
	"ja": "Japanese (日本語)",
}

// LanguageDirective returns the language instruction prefix to prepend to a
// prompt when CommunicationLanguage is set. Returns an empty string when lang
// is empty or unrecognized (falling back to the model's default language).
func LanguageDirective(lang string) string {
	if lang == "" {
		return ""
	}
	name, ok := languageNames[lang]
	if !ok {
		return ""
	}
	return "[IMPORTANT] You MUST respond in " + name + ". All your output must be in " + name + ".\n\n"
}
