package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

// agentJWTClaims is the registered + custom claims parsed from an agent JWT.
type agentJWTClaims struct {
	jwt.RegisteredClaims
	CompanyID   string   `json:"company_id"`
	AdapterType string   `json:"adapter_type"`
	RunID       string   `json:"run_id"`
	Permissions []string `json:"permissions,omitempty"`
}

// AgentJWTValidator validates HMAC-SHA256 signed agent JWTs.
type AgentJWTValidator struct {
	secret   []byte
	issuer   string
	audience string
}

// NewAgentJWTValidator creates a validator.
// If secret is empty it falls back to the SOKSAK_AGENT_JWT_SECRET env var.
func NewAgentJWTValidator(secret string) *AgentJWTValidator {
	if secret == "" {
		secret = os.Getenv("SOKSAK_AGENT_JWT_SECRET")
	}
	issuer := os.Getenv("SOKSAK_AGENT_JWT_ISSUER")
	if issuer == "" {
		issuer = "soksak"
	}
	audience := os.Getenv("SOKSAK_AGENT_JWT_AUDIENCE")
	if audience == "" {
		audience = "soksak-api"
	}
	return &AgentJWTValidator{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
	}
}

// Validate parses and verifies a JWT string, returning an Actor on success.
// Returns nil, nil if the token is structurally valid JWT but fails claims validation.
// Returns an error only for unexpected failures.
func (v *AgentJWTValidator) Validate(tokenStr string) (*Actor, error) {
	if len(v.secret) == 0 {
		return nil, nil
	}

	var claims agentJWTClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	},
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, nil
	}

	agentID := claims.Subject
	if agentID == "" || claims.CompanyID == "" {
		return nil, nil
	}

	return &Actor{
		Type:        ActorTypeAgent,
		ID:          agentID,
		AgentID:     agentID,
		CompanyID:   claims.CompanyID,
		RunID:       claims.RunID,
		Permissions: claims.Permissions,
		Source:      "agent_jwt",
	}, nil
}

// apiKeyRow holds the data returned from an API key lookup.
type apiKeyRow struct {
	ID        string `db:"id"`
	AgentID   string `db:"agent_id"`
	CompanyID string `db:"company_id"`
}

// ValidateAPIKey looks up a raw API key by its SHA-256 hash in the agent_api_keys table.
// Returns nil, nil if no matching active key is found.
func ValidateAPIKey(ctx context.Context, db *sqlx.DB, key string) (*Actor, error) {
	keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))

	var row apiKeyRow
	const query = `
		SELECT id, agent_id, company_id
		FROM   agent_api_keys
		WHERE  key_hash = $1
		  AND  revoked_at IS NULL
		LIMIT  1
	`
	err := db.GetContext(ctx, &row, query, keyHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("api key lookup: %w", err)
	}

	// Update last_used_at asynchronously — failure is non-fatal.
	_, _ = db.ExecContext(ctx,
		`UPDATE agent_api_keys SET last_used_at = NOW() WHERE id = $1`,
		row.ID,
	)

	return &Actor{
		Type:      ActorTypeAgent,
		ID:        row.AgentID,
		AgentID:   row.AgentID,
		CompanyID: row.CompanyID,
		KeyID:     row.ID,
		Source:    "agent_key",
	}, nil
}
