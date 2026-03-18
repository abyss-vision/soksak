package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"soksak/internal/middleware"
	"soksak/internal/services"
)

// IssueRoutes returns a chi.Router for issue endpoints.
// It must be mounted under a route that already provides {companyUuid}.
func IssueRoutes(svc *services.IssueService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", listIssues(svc))
	r.Post("/", createIssue(svc))

	r.Route("/{issueUuid}", func(r chi.Router) {
		r.Get("/", getIssue(svc))
		r.Patch("/", updateIssue(svc))
		r.Delete("/", deleteIssue(svc))
	})

	return r
}

func listIssues(svc *services.IssueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		issues, err := svc.List(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list issues"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issues)
	}
}

func getIssue(svc *services.IssueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		issueUUID := chi.URLParam(r, "issueUuid")
		issue, err := svc.Get(r.Context(), companyUUID, issueUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get issue"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
	}
}

func createIssue(svc *services.IssueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		input, err := middleware.BindAndValidate[services.CreateIssueInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		issue, err := svc.Create(r.Context(), companyUUID, *input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create issue"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(issue)
	}
}

func updateIssue(svc *services.IssueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		issueUUID := chi.URLParam(r, "issueUuid")
		input, err := middleware.BindAndValidate[services.UpdateIssueInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		issue, err := svc.Update(r.Context(), companyUUID, issueUUID, *input)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update issue"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
	}
}

func deleteIssue(svc *services.IssueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		issueUUID := chi.URLParam(r, "issueUuid")
		if err := svc.Delete(r.Context(), companyUUID, issueUUID); err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to delete issue"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}
