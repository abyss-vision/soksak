package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"abyss-view/internal/auth"
	"github.com/jmoiron/sqlx"
)

type contextKey string

const actorKey contextKey = "actor"

// AuthConfig carries configuration options for ActorMiddleware.
type AuthConfig struct {
	// DeploymentMode controls authentication behaviour.
	// When "local_trusted" all requests are granted a board_user actor.
	DeploymentMode string

	// JWTSecret is the HMAC-SHA256 secret for agent JWTs.
	// Falls back to the SOKSAK_AGENT_JWT_SECRET env var when empty.
	JWTSecret string
}

// ActorMiddleware resolves the authenticated actor for each request and stores it in the context.
//
// Resolution order:
//  1. X-Local-Trusted header present AND DeploymentMode == "local_trusted" → board_user
//  2. Authorization: Bearer <token> → try JWT (agent actor); if JWT invalid, try API key (agent actor)
//  3. Cookie session token → user actor via SessionStore
//
// On any successful resolution the actor is written to context; the request continues.
// On auth failure a 401 JSON response is written and the chain is halted.
// When no credentials are present at all the request continues with a "none" actor.
func ActorMiddleware(db *sqlx.DB, sessionStore *auth.SessionStore, jwtValidator *auth.AgentJWTValidator, cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1. Local-trusted short-circuit.
			if cfg.DeploymentMode == "local_trusted" {
				actor := &auth.Actor{
					Type:            auth.ActorTypeBoard,
					ID:              "local-board",
					UserID:          "local-board",
					IsInstanceAdmin: true,
					Source:          "local_implicit",
				}
				next.ServeHTTP(w, r.WithContext(withActor(ctx, actor)))
				return
			}

			// 2. Authorization: Bearer <token>
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				rawToken := strings.TrimSpace(authHeader[len("bearer "):])
				if rawToken == "" {
					writeUnauthorized(w)
					return
				}

				// Try JWT first.
				actor, err := jwtValidator.Validate(rawToken)
				if err != nil {
					writeUnauthorized(w)
					return
				}
				if actor != nil {
					// Optionally carry run-id from header.
					if rid := r.Header.Get("X-Soksak-Run-Id"); rid != "" {
						actor.RunID = rid
					}
					next.ServeHTTP(w, r.WithContext(withActor(ctx, actor)))
					return
				}

				// JWT failed — try API key (only when db is available).
				if db != nil {
					actor, err = auth.ValidateAPIKey(ctx, db, rawToken)
					if err != nil {
						writeUnauthorized(w)
						return
					}
				}
				if actor != nil {
					if rid := r.Header.Get("X-Soksak-Run-Id"); rid != "" {
						actor.RunID = rid
					}
					next.ServeHTTP(w, r.WithContext(withActor(ctx, actor)))
					return
				}

				// Bearer token present but unrecognised → 401.
				writeUnauthorized(w)
				return
			}

			// 3. Cookie session.
			if sessionStore != nil {
				sessionCookie, err := r.Cookie("better-auth.session_token")
				if err == nil && sessionCookie.Value != "" {
					actor, err := sessionStore.ValidateSession(ctx, sessionCookie.Value)
					if err != nil {
						writeUnauthorized(w)
						return
					}
					if actor != nil {
						next.ServeHTTP(w, r.WithContext(withActor(ctx, actor)))
						return
					}
				}
			}

			// No credentials — continue with none actor.
			noActor := &auth.Actor{Type: auth.ActorTypeNone, Source: "none"}
			next.ServeHTTP(w, r.WithContext(withActor(ctx, noActor)))
		})
	}
}

// ActorFromContext retrieves the Actor set by ActorMiddleware.
// Returns nil if no actor is present in the context.
func ActorFromContext(ctx context.Context) *auth.Actor {
	actor, _ := ctx.Value(actorKey).(*auth.Actor)
	return actor
}

func withActor(ctx context.Context, actor *auth.Actor) context.Context {
	return context.WithValue(ctx, actorKey, actor)
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "unauthorized",
		"code":  "UNAUTHORIZED",
	})
}
