package middleware

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

var safeMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

// BoardMutationGuard checks whether the company board targeted by the request
// is marked read-only in board_settings. If it is, POST/PATCH/DELETE requests
// are rejected with 405 Method Not Allowed.
//
// The {companyId} URL param (set by CompanyScope middleware) is used to look
// up the board's read-only flag. If no board_settings row exists the board is
// treated as writable (default-open policy).
func BoardMutationGuard(db *sqlx.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if safeMethods[r.Method] {
				next.ServeHTTP(w, r)
				return
			}

			companyID := CompanyFromContext(r.Context())
			if companyID == "" {
				// No company in scope — nothing to guard.
				next.ServeHTTP(w, r)
				return
			}

			readOnly, err := isBoardReadOnly(r, db, companyID)
			if err != nil {
				slog.Error("board_mutation_guard: failed to query board_settings",
					"company_id", companyID,
					"err", err,
				)
				// Fail open: don't block mutations on DB errors.
				next.ServeHTTP(w, r)
				return
			}

			if readOnly {
				WriteAppError(w, r, ErrMethodNotAllowed("Board is read-only"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isBoardReadOnly queries board_settings for the company. Returns false (writable)
// when no row exists or when the table does not yet exist.
func isBoardReadOnly(r *http.Request, db *sqlx.DB, companyID string) (bool, error) {
	var readOnly bool
	err := db.QueryRowContext(r.Context(),
		`SELECT read_only FROM board_settings WHERE company_id = $1 LIMIT 1`,
		companyID,
	).Scan(&readOnly)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		// Gracefully handle the case where the table doesn't exist yet.
		return false, nil
	}
	return readOnly, nil
}
