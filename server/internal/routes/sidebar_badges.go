package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"soksak/internal/middleware"
	"soksak/internal/services"
)

// SidebarBadgesHandler returns an http.HandlerFunc for GET /sidebar-badges.
// Must be used under a route that provides {companyUuid}.
func SidebarBadgesHandler(svc *services.DashboardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		badges, err := svc.GetSidebarBadges(r.Context(), companyUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get sidebar badges"))
			return
		}
		writeJSON(w, http.StatusOK, badges)
	}
}
