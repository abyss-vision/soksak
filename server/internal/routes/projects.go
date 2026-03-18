package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"abyss-view/internal/middleware"
	"abyss-view/internal/services"
)

// ProjectRoutes returns a chi.Router with CRUD and workspace routes for projects.
// Must be mounted under /api/companies/{companyUuid}/projects.
func ProjectRoutes(svc *services.ProjectService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		projects, err := svc.List(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list projects"))
			return
		}
		writeJSON(w, http.StatusOK, projects)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")

		var input services.CreateProjectInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		if input.Name == "" {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("name is required"))
			return
		}

		project, err := svc.Create(r.Context(), companyUUID, input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create project"))
			return
		}
		writeJSON(w, http.StatusCreated, project)
	})

	r.Get("/{projectUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		projectUUID := chi.URLParam(r, "projectUuid")

		project, err := svc.Get(r.Context(), companyUUID, projectUUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.WriteAppError(w, r, middleware.ErrNotFound("Project not found"))
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get project"))
			return
		}
		writeJSON(w, http.StatusOK, project)
	})

	r.Patch("/{projectUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		projectUUID := chi.URLParam(r, "projectUuid")

		var input services.UpdateProjectInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		project, err := svc.Update(r.Context(), companyUUID, projectUUID, input)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.WriteAppError(w, r, middleware.ErrNotFound("Project not found"))
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update project"))
			return
		}
		writeJSON(w, http.StatusOK, project)
	})

	r.Delete("/{projectUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		projectUUID := chi.URLParam(r, "projectUuid")

		if err := svc.Delete(r.Context(), companyUUID, projectUUID); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to delete project"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Get("/{projectUuid}/workspaces", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		projectUUID := chi.URLParam(r, "projectUuid")

		// Verify project exists within company first.
		if _, err := svc.Get(r.Context(), companyUUID, projectUUID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.WriteAppError(w, r, middleware.ErrNotFound("Project not found"))
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get project"))
			return
		}

		workspaces, err := svc.ListWorkspaces(r.Context(), companyUUID, projectUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list workspaces"))
			return
		}
		writeJSON(w, http.StatusOK, workspaces)
	})

	return r
}
