package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"abyss-view/internal/middleware"
	"abyss-view/internal/services"
)

// AccessRoutes returns a chi.Router for permission grant CRUD.
// Must be mounted under /api/companies/{companyUuid}/access.
func AccessRoutes(svc *services.AccessControlService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := middleware.CompanyFromContext(r.Context())
		grants, err := svc.ListGrants(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list permission grants"))
			return
		}
		writeJSON(w, http.StatusOK, grants)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := middleware.CompanyFromContext(r.Context())

		var input services.GrantPermissionInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		grant, err := svc.GrantPermission(r.Context(), companyUUID, input)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to grant permission"))
			return
		}
		writeJSON(w, http.StatusCreated, grant)
	})

	r.Delete("/{grantUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := middleware.CompanyFromContext(r.Context())
		grantUUID := chi.URLParam(r, "grantUuid")

		if err := svc.RevokePermission(r.Context(), companyUUID, grantUUID); err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to revoke permission"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
