package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"soksak/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

// agentJWTClaimsForTest matches the unexported struct in auth package — we build tokens manually.
type agentJWTClaimsForTest struct {
	jwt.RegisteredClaims
	CompanyID   string `json:"company_id"`
	AdapterType string `json:"adapter_type"`
	RunID       string `json:"run_id"`
}

func makeJWT(t *testing.T, secret, agentID, companyID, issuer, audience string, exp time.Time) string {
	t.Helper()
	claims := agentJWTClaimsForTest{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   agentID,
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		CompanyID:   companyID,
		AdapterType: "claude_local",
		RunID:       "run-1",
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return tok
}

func newMiddleware(deploymentMode string, secret string) func(http.Handler) http.Handler {
	validator := auth.NewAgentJWTValidator(secret)
	return ActorMiddleware(nil, nil, validator, AuthConfig{
		DeploymentMode: deploymentMode,
		JWTSecret:      secret,
	})
}

func captureActor(t *testing.T, mw func(http.Handler) http.Handler, req *http.Request) (*auth.Actor, int) {
	t.Helper()
	var captured *auth.Actor
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = ActorFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return captured, rr.Code
}

func TestActorMiddleware_LocalTrusted(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	mw := newMiddleware("local_trusted", "secret")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	actor, code := captureActor(t, mw, req)

	if code != http.StatusOK {
		t.Errorf("status = %d, want 200", code)
	}
	if actor == nil {
		t.Fatal("expected actor, got nil")
	}
	if actor.Type != auth.ActorTypeBoard {
		t.Errorf("type = %v, want board_user", actor.Type)
	}
	if !actor.IsInstanceAdmin {
		t.Error("expected IsInstanceAdmin=true")
	}
}

func TestActorMiddleware_ValidJWT(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	mw := newMiddleware("authenticated", "test-secret")

	tok := makeJWT(t, "test-secret", "agent-1", "company-1", "soksak", "soksak-api", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	actor, code := captureActor(t, mw, req)
	if code != http.StatusOK {
		t.Errorf("status = %d, want 200", code)
	}
	if actor == nil {
		t.Fatal("expected actor, got nil")
	}
	if actor.Type != auth.ActorTypeAgent {
		t.Errorf("type = %v, want agent", actor.Type)
	}
	if actor.AgentID != "agent-1" {
		t.Errorf("AgentID = %q, want agent-1", actor.AgentID)
	}
}

func TestActorMiddleware_InvalidBearerToken_Returns401(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	mw := newMiddleware("authenticated", "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer notavalidtoken")

	_, code := captureActor(t, mw, req)
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", code)
	}
}

func TestActorMiddleware_NoCredentials_NoneActor(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	mw := newMiddleware("authenticated", "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	actor, code := captureActor(t, mw, req)

	if code != http.StatusOK {
		t.Errorf("status = %d, want 200", code)
	}
	if actor == nil {
		t.Fatal("expected actor, got nil")
	}
	if actor.Type != auth.ActorTypeNone {
		t.Errorf("type = %v, want none", actor.Type)
	}
}

func TestActorMiddleware_RunIDPropagated(t *testing.T) {
	t.Setenv("SOKSAK_AGENT_JWT_ISSUER", "soksak")
	t.Setenv("SOKSAK_AGENT_JWT_AUDIENCE", "soksak-api")
	mw := newMiddleware("authenticated", "test-secret")

	tok := makeJWT(t, "test-secret", "agent-1", "company-1", "soksak", "soksak-api", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("X-Soksak-Run-Id", "run-override")

	actor, _ := captureActor(t, mw, req)
	if actor == nil {
		t.Fatal("expected actor")
	}
	if actor.RunID != "run-override" {
		t.Errorf("RunID = %q, want %q", actor.RunID, "run-override")
	}
}
