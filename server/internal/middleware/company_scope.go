package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type companyScopeKey string

const companyKey companyScopeKey = "company_id"

// CompanyFromContext returns the company ID stored in the context by CompanyScope.
func CompanyFromContext(ctx context.Context) string {
	id, _ := ctx.Value(companyKey).(string)
	return id
}

// CompanyScope extracts the {companyUuid} Chi URL parameter and verifies the
// requesting actor is an active member of that company. On success the
// company UUID is stored in the request context. On failure a 403 is returned.
//
// This middleware depends on ActorMiddleware having already run.
func CompanyScope(db *sqlx.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			companyID := chi.URLParam(r, "companyUuid")
			if companyID == "" {
				WriteAppError(w, r, ErrForbidden("Company UUID required"))
				return
			}

			actor := ActorFromContext(r.Context())
			if actor == nil {
				WriteAppError(w, r, ErrForbidden("Authentication required"))
				return
			}

			// Instance admins bypass membership checks.
			if actor.IsInstanceAdmin {
				ctx := context.WithValue(r.Context(), companyKey, companyID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			isMember, err := isCompanyMember(r, db, actor.ID, companyID)
			if err != nil {
				WriteAppError(w, r, ErrInternal("Failed to verify company membership"))
				return
			}
			if !isMember {
				WriteAppError(w, r, ErrForbidden("Not a member of this company"))
				return
			}

			ctx := context.WithValue(r.Context(), companyKey, companyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isCompanyMember checks whether principalID has an active membership in the company.
func isCompanyMember(r *http.Request, db *sqlx.DB, principalID, companyID string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(r.Context(),
		`SELECT EXISTS(
			SELECT 1 FROM company_memberships
			WHERE principal_id  = $1
			  AND company_uuid  = $2
			  AND status        = 'active'
		)`,
		principalID,
		companyID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
