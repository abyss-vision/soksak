package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"abyss-view/internal/middleware"
	"abyss-view/internal/services"
)

// GoalRoutes returns a chi.Router for goal endpoints.
// It must be mounted under a route that already provides {companyUuid}.
func GoalRoutes(svc *services.GoalService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", listGoals(svc))
	r.Post("/", createGoal(svc))

	r.Route("/{goalUuid}", func(r chi.Router) {
		r.Get("/", getGoal(svc))
		r.Patch("/", updateGoal(svc))
		r.Delete("/", deleteGoal(svc))
		r.Get("/issues", listGoalIssues(svc))
	})

	return r
}

func listGoals(svc *services.GoalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		goals, err := svc.List(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list goals"))
			return
		}
		writeJSON(w, http.StatusOK, goals)
	}
}

func getGoal(svc *services.GoalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		goalUUID := chi.URLParam(r, "goalUuid")
		goal, err := svc.Get(r.Context(), companyUUID, goalUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get goal"))
			return
		}
		writeJSON(w, http.StatusOK, goal)
	}
}

func createGoal(svc *services.GoalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		input, err := middleware.BindAndValidate[services.CreateGoalInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		goal, err := svc.Create(r.Context(), companyUUID, *input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create goal"))
			return
		}
		writeJSON(w, http.StatusCreated, goal)
	}
}

func updateGoal(svc *services.GoalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		goalUUID := chi.URLParam(r, "goalUuid")
		input, err := middleware.BindAndValidate[services.UpdateGoalInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		goal, err := svc.Update(r.Context(), companyUUID, goalUUID, *input)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update goal"))
			return
		}
		writeJSON(w, http.StatusOK, goal)
	}
}

func deleteGoal(svc *services.GoalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		goalUUID := chi.URLParam(r, "goalUuid")
		if err := svc.Delete(r.Context(), companyUUID, goalUUID); err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to delete goal"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func listGoalIssues(svc *services.GoalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		goalUUID := chi.URLParam(r, "goalUuid")
		issues, err := svc.ListLinkedIssues(r.Context(), companyUUID, goalUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list goal issues"))
			return
		}
		writeJSON(w, http.StatusOK, issues)
	}
}
