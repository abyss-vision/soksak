package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"soksak/internal/middleware"
	"soksak/internal/services"
)

// CostRoutes returns a chi.Router for cost and budget endpoints.
// Must be mounted under a route that provides {companyUuid}.
func CostRoutes(costSvc *services.CostService, budgetSvc *services.BudgetService) chi.Router {
	r := chi.NewRouter()

	// Cost events.
	r.Get("/", listCosts(costSvc))
	r.Post("/", recordCost(costSvc))
	r.Get("/summary", getCostSummary(costSvc))
	r.Get("/by-agent", getCostsByAgent(costSvc))

	// Budget policies sub-resource.
	r.Route("/policies", func(r chi.Router) {
		r.Get("/", listBudgetPolicies(budgetSvc))
		r.Post("/", createBudgetPolicy(budgetSvc))
		r.Route("/{policyUuid}", func(r chi.Router) {
			r.Get("/", getBudgetPolicy(budgetSvc))
			r.Patch("/", updateBudgetPolicy(budgetSvc))
			r.Delete("/", deleteBudgetPolicy(budgetSvc))
		})
	})

	// Budget incidents.
	r.Get("/incidents", listBudgetIncidents(budgetSvc))

	return r
}

func parseDateParam(r *http.Request, key string) (*time.Time, *middleware.AppError) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		appErr := middleware.ErrUnprocessable("invalid '" + key + "' date, expected RFC3339")
		return nil, appErr
	}
	return &t, nil
}

func parseLimitParam(r *http.Request) (int, *middleware.AppError) {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return 100, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 || n > 500 {
		return 0, middleware.ErrUnprocessable("invalid 'limit' value")
	}
	return n, nil
}

func parseOffsetParam(r *http.Request) int {
	raw := r.URL.Query().Get("offset")
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func listCosts(svc *services.CostService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		limit, appErr := parseLimitParam(r)
		if appErr != nil {
			middleware.WriteAppError(w, r, appErr)
			return
		}
		events, err := svc.ListCosts(r.Context(), companyUUID, limit)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list costs"))
			return
		}
		writeJSON(w, http.StatusOK, events)
	}
}

func recordCost(svc *services.CostService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		input, err := middleware.BindAndValidate[services.RecordCostInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		event, err := svc.RecordCost(r.Context(), companyUUID, *input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to record cost"))
			return
		}
		writeJSON(w, http.StatusCreated, event)
	}
}

func getCostSummary(svc *services.CostService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		from, appErr := parseDateParam(r, "from")
		if appErr != nil {
			middleware.WriteAppError(w, r, appErr)
			return
		}
		to, appErr := parseDateParam(r, "to")
		if appErr != nil {
			middleware.WriteAppError(w, r, appErr)
			return
		}
		summary, err := svc.GetCostSummary(r.Context(), companyUUID, from, to)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get cost summary"))
			return
		}
		writeJSON(w, http.StatusOK, summary)
	}
}

func getCostsByAgent(svc *services.CostService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		from, appErr := parseDateParam(r, "from")
		if appErr != nil {
			middleware.WriteAppError(w, r, appErr)
			return
		}
		to, appErr := parseDateParam(r, "to")
		if appErr != nil {
			middleware.WriteAppError(w, r, appErr)
			return
		}
		rows, err := svc.GetCostsByAgent(r.Context(), companyUUID, from, to)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get costs by agent"))
			return
		}
		writeJSON(w, http.StatusOK, rows)
	}
}

func listBudgetPolicies(svc *services.BudgetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		policies, err := svc.ListBudgetPolicies(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list budget policies"))
			return
		}
		writeJSON(w, http.StatusOK, policies)
	}
}

func getBudgetPolicy(svc *services.BudgetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		policyUUID := chi.URLParam(r, "policyUuid")
		policy, err := svc.GetBudgetPolicy(r.Context(), companyUUID, policyUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get budget policy"))
			return
		}
		writeJSON(w, http.StatusOK, policy)
	}
}

func createBudgetPolicy(svc *services.BudgetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		input, err := middleware.BindAndValidate[services.CreateBudgetPolicyInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		policy, err := svc.CreateBudgetPolicy(r.Context(), companyUUID, *input, nil)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create budget policy"))
			return
		}
		writeJSON(w, http.StatusCreated, policy)
	}
}

func updateBudgetPolicy(svc *services.BudgetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		policyUUID := chi.URLParam(r, "policyUuid")
		input, err := middleware.BindAndValidate[services.UpdateBudgetPolicyInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		policy, err := svc.UpdateBudgetPolicy(r.Context(), companyUUID, policyUUID, *input, nil)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update budget policy"))
			return
		}
		writeJSON(w, http.StatusOK, policy)
	}
}

func deleteBudgetPolicy(svc *services.BudgetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		policyUUID := chi.URLParam(r, "policyUuid")
		if err := svc.DeleteBudgetPolicy(r.Context(), companyUUID, policyUUID); err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to delete budget policy"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func listBudgetIncidents(svc *services.BudgetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		incidents, err := svc.ListBudgetIncidents(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list budget incidents"))
			return
		}
		writeJSON(w, http.StatusOK, incidents)
	}
}
