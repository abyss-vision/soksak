package routes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"soksak/internal/middleware"
	"soksak/internal/services"
)

// AgentRoutes returns a chi.Router with CRUD and lifecycle routes for agents.
// Must be mounted under /api/companies/{companyUuid}/agents.
func AgentRoutes(svc *services.AgentService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		agents, err := svc.List(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list agents"))
			return
		}
		writeJSON(w, http.StatusOK, agents)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")

		var input services.CreateAgentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		if input.Name == "" || input.Role == "" || input.AdapterType == "" {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("name, role, and adapterType are required"))
			return
		}

		agent, err := svc.Create(r.Context(), companyUUID, input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create agent"))
			return
		}
		writeJSON(w, http.StatusCreated, agent)
	})

	r.Get("/{agentUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		agentUUID := chi.URLParam(r, "agentUuid")

		agent, err := svc.Get(r.Context(), companyUUID, agentUUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.WriteAppError(w, r, middleware.ErrNotFound("Agent not found"))
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get agent"))
			return
		}
		writeJSON(w, http.StatusOK, agent)
	})

	r.Patch("/{agentUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		agentUUID := chi.URLParam(r, "agentUuid")

		var input services.UpdateAgentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		agent, err := svc.Update(r.Context(), companyUUID, agentUUID, input)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.WriteAppError(w, r, middleware.ErrNotFound("Agent not found"))
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update agent"))
			return
		}
		writeJSON(w, http.StatusOK, agent)
	})

	r.Delete("/{agentUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		agentUUID := chi.URLParam(r, "agentUuid")

		if err := svc.Delete(r.Context(), companyUUID, agentUUID); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to delete agent"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Post("/{agentUuid}/hire", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		agentUUID := chi.URLParam(r, "agentUuid")

		var body struct {
			Config json.RawMessage `json:"config"`
		}
		// Body is optional; ignore decode error.
		_ = json.NewDecoder(r.Body).Decode(&body)

		if err := svc.Hire(r.Context(), companyUUID, agentUUID, body.Config); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.WriteAppError(w, r, middleware.ErrNotFound("Agent not found"))
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to hire agent"))
			return
		}

		agent, err := svc.Get(r.Context(), companyUUID, agentUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to fetch agent after hire"))
			return
		}
		writeJSON(w, http.StatusOK, agent)
	})

	r.Post("/{agentUuid}/fire", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		agentUUID := chi.URLParam(r, "agentUuid")

		if err := svc.Fire(r.Context(), companyUUID, agentUUID); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to fire agent"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Post("/{agentUuid}/pause", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		agentUUID := chi.URLParam(r, "agentUuid")

		if err := svc.Pause(r.Context(), companyUUID, agentUUID); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to pause agent"))
			return
		}

		agent, err := svc.Get(r.Context(), companyUUID, agentUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to fetch agent after pause"))
			return
		}
		writeJSON(w, http.StatusOK, agent)
	})

	return r
}
