package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// SessionStore performs session CRUD against the better-auth compatible `session` table.
type SessionStore struct {
	db *sqlx.DB
}

// NewSessionStore creates a SessionStore backed by the given database.
func NewSessionStore(db *sqlx.DB) *SessionStore {
	return &SessionStore{db: db}
}

// sessionRow holds the joined data returned from a session lookup.
type sessionRow struct {
	UserID string `db:"user_id"`
}

// ValidateSession resolves a raw session token into an Actor.
// It SHA-256 hashes the raw token and looks it up in the `session` table (better-auth schema).
// Returns nil, nil if no matching active session is found.
func (s *SessionStore) ValidateSession(ctx context.Context, token string) (*Actor, error) {
	tokenHash := hashToken(token)

	var row sessionRow
	const query = `
		SELECT user_id
		FROM   session
		WHERE  token = $1
		  AND  expires_at > NOW()
		LIMIT  1
	`
	err := s.db.GetContext(ctx, &row, query, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("session lookup: %w", err)
	}

	return &Actor{
		Type:   ActorTypeUser,
		ID:     row.UserID,
		UserID: row.UserID,
		Source: "session",
	}, nil
}

// CreateSession inserts a new session row and returns the raw (unhashed) token.
// The token is a 32-byte cryptographically random hex string; only its SHA-256 hash is stored.
// ttl controls how long the session is valid.
func (s *SessionStore) CreateSession(ctx context.Context, userID string, ttl time.Duration) (rawToken string, err error) {
	raw, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	tokenHash := hashToken(raw)
	id := tokenHash // use hash as stable ID, consistent with better-auth convention
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)

	const query = `
		INSERT INTO session (id, token, user_id, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := s.db.ExecContext(ctx, query, id, tokenHash, userID, expiresAt, now, now); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return raw, nil
}

// RevokeSession deletes the session matching the given raw token.
// Returns nil if the session did not exist (idempotent).
func (s *SessionStore) RevokeSession(ctx context.Context, token string) error {
	tokenHash := hashToken(token)
	const query = `DELETE FROM session WHERE token = $1`
	if _, err := s.db.ExecContext(ctx, query, tokenHash); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// RevokeAllUserSessions deletes every active session for the given user.
func (s *SessionStore) RevokeAllUserSessions(ctx context.Context, userID string) error {
	const query = `DELETE FROM session WHERE user_id = $1`
	if _, err := s.db.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

// hashToken returns the SHA-256 hex digest of the raw token, matching better-auth's storage format.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

// generateToken creates a 32-byte cryptographically random hex string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
