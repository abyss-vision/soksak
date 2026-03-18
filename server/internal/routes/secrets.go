package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"soksak/internal/middleware"
	"soksak/internal/services"
)

// SecretRoutes returns a chi.Router with CRUD endpoints for company secrets.
// Values are NEVER returned in any GET response.
func SecretRoutes(svc *services.SecretService) chi.Router {
	r := chi.NewRouter()

	// GET /secrets — list all secret metadata for the company
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		items, err := svc.List(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list secrets"))
			return
		}
		writeJSON(w, http.StatusOK, items)
	})

	// GET /secrets/{secretUuid} — get single secret metadata
	r.Get("/{secretUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		secretUUID := chi.URLParam(r, "secretUuid")

		item, err := svc.Get(r.Context(), companyUUID, secretUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get secret"))
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	// POST /secrets — create a new secret
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")

		var input services.CreateSecretInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		item, err := svc.Create(r.Context(), companyUUID, input)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create secret"))
			return
		}
		writeJSON(w, http.StatusCreated, item)
	})

	// PATCH /secrets/{secretUuid} — update secret metadata
	r.Patch("/{secretUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		secretUUID := chi.URLParam(r, "secretUuid")

		var input services.UpdateSecretInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		item, err := svc.Update(r.Context(), companyUUID, secretUUID, input)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update secret"))
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	// DELETE /secrets/{secretUuid} — permanently delete a secret
	r.Delete("/{secretUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		secretUUID := chi.URLParam(r, "secretUuid")

		if err := svc.Delete(r.Context(), companyUUID, secretUUID); err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to delete secret"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
