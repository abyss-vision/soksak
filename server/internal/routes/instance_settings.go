package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"abyss-view/internal/middleware"
	"abyss-view/internal/services"
)

// InstanceSettingsRoutes returns a chi.Router for GET/PATCH /api/instance-settings.
func InstanceSettingsRoutes(svc *services.InstanceSettingsService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		settings, err := svc.Get(r.Context())
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get instance settings"))
			return
		}
		writeJSON(w, http.StatusOK, settings)
	})

	r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
		var patch map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		updated, err := svc.Update(r.Context(), patch)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update instance settings"))
			return
		}
		writeJSON(w, http.StatusOK, updated)
	})

	return r
}
