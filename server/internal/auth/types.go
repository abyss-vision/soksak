package auth

// ActorType identifies what kind of principal is making a request.
type ActorType string

const (
	ActorTypeBoard    ActorType = "board_user"
	ActorTypeAgent    ActorType = "agent"
	ActorTypeAPIKey   ActorType = "api_key"
	ActorTypeUser     ActorType = "user"
	ActorTypeNone     ActorType = "none"
)

// Actor represents the authenticated principal for a request.
type Actor struct {
	Type      ActorType
	ID        string
	CompanyID string

	// AgentID is set when Type == ActorTypeAgent.
	AgentID string
	// KeyID is set when Type == ActorTypeAPIKey.
	KeyID string
	// UserID is set when Type == ActorTypeUser or ActorTypeBoard.
	UserID string
	// RunID is an optional identifier for the current run/session.
	RunID string
	// IsInstanceAdmin is true for board_user actors with instance-admin rights.
	IsInstanceAdmin bool
	// Permissions holds any explicit permission strings carried in the JWT claims.
	Permissions []string

	// Source describes how the actor was resolved (e.g. "session", "agent_jwt", "agent_key", "local_implicit").
	Source string
}

// Claims is the custom JWT claims payload for agent tokens.
type Claims struct {
	Sub         string   `json:"sub"`
	CompanyID   string   `json:"company_id"`
	AdapterType string   `json:"adapter_type"`
	RunID       string   `json:"run_id"`
	Permissions []string `json:"permissions,omitempty"`
	Issuer      string   `json:"iss,omitempty"`
	Audience    string   `json:"aud,omitempty"`
}
