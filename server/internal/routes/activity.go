package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"abyss-view/internal/middleware"
	"abyss-view/internal/services"
)

// ActivityRoutes returns a chi.Router for activity log endpoints.
// Must be mounted under a route that provides {companyUuid}.
func ActivityRoutes(svc *services.ActivityService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", listActivity(svc))
	r.Post("/", recordActivity(svc))

	return r
}

func listActivity(svc *services.ActivityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")

		limit, appErr := parseLimitParam(r)
		if appErr != nil {
			middleware.WriteAppError(w, r, appErr)
			return
		}
		offset := parseOffsetParam(r)

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

		var agentUUID *string
		if v := r.URL.Query().Get("agentUuid"); v != "" {
			agentUUID = &v
		}
		var entityType *string
		if v := r.URL.Query().Get("entityType"); v != "" {
			entityType = &v
		}
		var entityID *string
		if v := r.URL.Query().Get("entityId"); v != "" {
			entityID = &v
		}

		filters := services.ActivityFilters{
			AgentUUID:  agentUUID,
			EntityType: entityType,
			EntityID:   entityID,
			From:       from,
			To:         to,
			Limit:      limit,
			Offset:     offset,
		}

		entries, err := svc.ListActivity(r.Context(), companyUUID, filters)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list activity"))
			return
		}
		writeJSON(w, http.StatusOK, entries)
	}
}

func recordActivity(svc *services.ActivityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		input, err := middleware.BindAndValidate[services.RecordActivityInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		entry, err := svc.RecordActivity(r.Context(), companyUUID, *input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to record activity"))
			return
		}
		writeJSON(w, http.StatusCreated, entry)
	}
}
