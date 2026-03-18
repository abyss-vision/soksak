package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"abyss-view/internal/middleware"
	"abyss-view/internal/services"
)

// ApprovalRoutes returns a chi.Router for approval endpoints.
// It must be mounted under a route that already provides {companyUuid}.
func ApprovalRoutes(svc *services.ApprovalService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", listApprovals(svc))
	r.Post("/", createApproval(svc))

	r.Route("/{approvalUuid}", func(r chi.Router) {
		r.Get("/", getApproval(svc))
		r.Post("/resolve", resolveApproval(svc))
	})

	return r
}

func listApprovals(svc *services.ApprovalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		status := r.URL.Query().Get("status")
		approvals, err := svc.List(r.Context(), companyUUID, status)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list approvals"))
			return
		}
		writeJSON(w, http.StatusOK, approvals)
	}
}

func getApproval(svc *services.ApprovalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		approvalUUID := chi.URLParam(r, "approvalUuid")
		approval, err := svc.Get(r.Context(), companyUUID, approvalUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get approval"))
			return
		}
		writeJSON(w, http.StatusOK, approval)
	}
}

func createApproval(svc *services.ApprovalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		input, err := middleware.BindAndValidate[services.CreateApprovalInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		approval, err := svc.Create(r.Context(), companyUUID, *input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create approval"))
			return
		}
		writeJSON(w, http.StatusCreated, approval)
	}
}

// resolveApprovalRequest is the body for the POST /{approvalUuid}/resolve endpoint.
type resolveApprovalRequest struct {
	Action          string  `json:"action"          validate:"required,oneof=approve reject"`
	DecisionNote    *string `json:"decisionNote"`
	DecidedByUserID *string `json:"decidedByUserId"`
}

func resolveApproval(svc *services.ApprovalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		approvalUUID := chi.URLParam(r, "approvalUuid")

		input, err := middleware.BindAndValidate[resolveApprovalRequest](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		targetStatus := "approved"
		if input.Action == "reject" {
			targetStatus = "rejected"
		}

		resolveInput := services.ResolveApprovalInput{
			DecisionNote:    input.DecisionNote,
			DecidedByUserID: input.DecidedByUserID,
		}

		approval, err := svc.Resolve(r.Context(), companyUUID, approvalUUID, targetStatus, resolveInput)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to resolve approval"))
			return
		}
		writeJSON(w, http.StatusOK, approval)
	}
}
