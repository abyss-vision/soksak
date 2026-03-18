package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"soksak/internal/middleware"
	"soksak/internal/services"
)

// CompanyRoutes returns a chi.Router with all company-related endpoints.
func CompanyRoutes(svc *services.CompanyService) chi.Router {
	r := chi.NewRouter()

	r.Get("/", listCompanies(svc))
	r.Post("/", createCompany(svc))

	r.Route("/{companyUuid}", func(r chi.Router) {
		r.Get("/", getCompany(svc))
		r.Patch("/", updateCompany(svc))
		r.Delete("/", deleteCompany(svc))
		r.Get("/members", listMembers(svc))
	})

	return r
}

func listCompanies(svc *services.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companies, err := svc.List(r.Context())
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list companies"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(companies)
	}
}

func getCompany(svc *services.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		company, err := svc.Get(r.Context(), companyUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get company"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(company)
	}
}

func createCompany(svc *services.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := middleware.BindAndValidate[services.CreateCompanyInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		company, err := svc.Create(r.Context(), *input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create company"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(company)
	}
}

func updateCompany(svc *services.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		input, err := middleware.BindAndValidate[services.UpdateCompanyInput](r)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		company, err := svc.Update(r.Context(), companyUUID, *input)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update company"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(company)
	}
}

func deleteCompany(svc *services.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		if err := svc.Delete(r.Context(), companyUUID); err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to delete company"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func listMembers(svc *services.CompanyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		members, err := svc.ListMembers(r.Context(), companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list members"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(members)
	}
}
