package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"soksak/internal/middleware"
	"soksak/internal/services"
)

// DashboardRoutes returns a chi.Router for dashboard endpoints.
// Must be mounted under a route that provides {companyUuid}.
func DashboardRoutes(svc *services.DashboardService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", getDashboard(svc))

	return r
}

func getDashboard(svc *services.DashboardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		summary, err := svc.GetDashboard(r.Context(), companyUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get dashboard"))
			return
		}
		writeJSON(w, http.StatusOK, summary)
	}
}
