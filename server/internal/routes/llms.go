package routes

import (
	"net/http"
)

// LLMsHandler returns a handler for GET /llms that stubs an empty model list.
// The adapter registry is not yet wired; this stub satisfies the route contract.
func LLMsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []interface{}{})
	}
}
