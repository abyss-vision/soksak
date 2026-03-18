package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func makeTestToken(t *testing.T, secret string, claims agentJWTClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestAgentJWTValidator_Valid(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	v := NewAgentJWTValidator("test-secret")

	claims := agentJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "agent-1",
			Issuer:    "soksak",
			Audience:  jwt.ClaimStrings{"soksak-api"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		CompanyID:   "company-1",
		AdapterType: "claude_local",
		RunID:       "run-1",
	}
	tok := makeTestToken(t, "test-secret", claims)

	actor, err := v.Validate(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor == nil {
		t.Fatal("expected actor, got nil")
	}
	if actor.AgentID != "agent-1" {
		t.Errorf("AgentID = %q, want %q", actor.AgentID, "agent-1")
	}
	if actor.CompanyID != "company-1" {
		t.Errorf("CompanyID = %q, want %q", actor.CompanyID, "company-1")
	}
	if actor.Source != "agent_jwt" {
		t.Errorf("Source = %q, want %q", actor.Source, "agent_jwt")
	}
	if actor.Type != ActorTypeAgent {
		t.Errorf("Type = %v, want %v", actor.Type, ActorTypeAgent)
	}
}

func TestAgentJWTValidator_Expired(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	v := NewAgentJWTValidator("test-secret")

	claims := agentJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "agent-1",
			Issuer:    "soksak",
			Audience:  jwt.ClaimStrings{"soksak-api"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
		CompanyID: "company-1",
	}
	tok := makeTestToken(t, "test-secret", claims)

	actor, err := v.Validate(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor != nil {
		t.Errorf("expected nil actor for expired token, got %+v", actor)
	}
}

func TestAgentJWTValidator_WrongSecret(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	v := NewAgentJWTValidator("real-secret")

	claims := agentJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "agent-1",
			Issuer:    "soksak",
			Audience:  jwt.ClaimStrings{"soksak-api"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		CompanyID: "company-1",
	}
	tok := makeTestToken(t, "wrong-secret", claims)

	actor, err := v.Validate(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor != nil {
		t.Errorf("expected nil actor for wrong secret, got %+v", actor)
	}
}

func TestAgentJWTValidator_EmptySecret(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_SECRET", "")
	v := NewAgentJWTValidator("")
	actor, err := v.Validate("any.token.here")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor != nil {
		t.Errorf("expected nil when secret is empty, got %+v", actor)
	}
}

func TestAgentJWTValidator_IssuerMismatch(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	v := NewAgentJWTValidator("test-secret")

	claims := agentJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "agent-1",
			Issuer:    "other-issuer",
			Audience:  jwt.ClaimStrings{"soksak-api"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		CompanyID: "company-1",
	}
	tok := makeTestToken(t, "test-secret", claims)

	actor, err := v.Validate(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actor != nil {
		t.Errorf("expected nil actor for issuer mismatch, got %+v", actor)
	}
}
