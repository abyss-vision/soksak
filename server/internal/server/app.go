package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	apii18n "abyss-view/internal/i18n"
	"abyss-view/internal/adapters/process"
	"abyss-view/internal/auth"
	"abyss-view/internal/db"
	"abyss-view/internal/middleware"
	"abyss-view/internal/realtime"
	"abyss-view/internal/routes"
	"abyss-view/internal/services"
	"abyss-view/internal/storage"
)

// App holds all application dependencies and the HTTP router.
type App struct {
	Config  *Config
	DB      *sqlx.DB
	Bundle  *i18n.Bundle
	Router  chi.Router
	Hub     *realtime.Hub
	PM      *process.Manager
}

// NewApp creates a new App, wires middleware, and sets up routes.
func NewApp(cfg *Config, database *sqlx.DB, bundle *i18n.Bundle) *App {
	return NewAppWithDeps(cfg, database, bundle, nil, nil)
}

// NewAppWithDeps creates a new App with optional hub and process manager.
func NewAppWithDeps(cfg *Config, database *sqlx.DB, bundle *i18n.Bundle, hub *realtime.Hub, pm *process.Manager) *App {
	if bundle == nil {
		bundle = apii18n.NewBundle(language.English)
	}
	if hub == nil {
		hub = realtime.NewHub()
	}
	if pm == nil {
		pm = process.New()
	}

	sessionStore := auth.NewSessionStore(database)
	jwtValidator := auth.NewAgentJWTValidator(cfg.JWTSecret)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.ErrorHandler)
	r.Use(apii18n.LocaleMiddleware(bundle))
	r.Use(middleware.ActorMiddleware(database, sessionStore, jwtValidator, middleware.AuthConfig{
		DeploymentMode: cfg.DeploymentMode,
		JWTSecret:      cfg.JWTSecret,
	}))

	app := &App{
		Config: cfg,
		DB:     database,
		Bundle: bundle,
		Router: r,
		Hub:    hub,
		PM:     pm,
	}
	app.SetupRoutes()
	return app
}

// SetupRoutes mounts all route groups onto the router.
func (a *App) SetupRoutes() {
	a.Router.Get("/api/health", a.handleHealth)

	if a.DB == nil {
		return
	}

	// Company-scoped subrouter with CompanyScope middleware.
	companySvc := services.NewCompanyService(a.DB)
	agentSvc := services.NewAgentService(a.DB)
	projectSvc := services.NewProjectService(a.DB)
	issueSvc := services.NewIssueService(a.DB, a.Hub)
	goalSvc := services.NewGoalService(a.DB)
	approvalSvc := services.NewApprovalService(a.DB, a.Hub)
	settingsSvc := services.NewInstanceSettingsService(a.DB)
	accessSvc := services.NewAccessControlService(a.DB)
	costSvc := services.NewCostService(a.DB)
	budgetSvc := services.NewBudgetService(a.DB)
	activitySvc := services.NewActivityService(a.DB)
	dashboardSvc := services.NewDashboardService(a.DB)
	secretSvc := services.NewSecretService(a.DB)
	assetSvc := services.NewAssetService(a.DB)
	storageSvc := storage.NewLocalDiskProvider(a.Config.StorageBaseDir)

	// Top-level company routes.
	a.Router.Mount("/api/companies", routes.CompanyRoutes(companySvc))

	// Instance settings at top level (not company-scoped).
	a.Router.Mount("/api/instance-settings", routes.InstanceSettingsRoutes(settingsSvc))

	// Company-scoped sub-routes.
	a.Router.Route("/api/companies/{companyUuid}", func(r chi.Router) {
		r.Use(middleware.CompanyScope(a.DB))

		r.Mount("/agents", routes.AgentRoutes(agentSvc))
		r.Mount("/issues", routes.IssueRoutes(issueSvc))
		r.Mount("/projects", routes.ProjectRoutes(projectSvc))
		r.Mount("/goals", routes.GoalRoutes(goalSvc))
		r.Mount("/approvals", routes.ApprovalRoutes(approvalSvc))
		r.Mount("/access", routes.AccessRoutes(accessSvc))
		r.Mount("/execution-workspaces", routes.ExecutionWorkspaceRoutes(a.DB))
		r.Get("/llms", routes.LLMsHandler())
		r.Mount("/costs", routes.CostRoutes(costSvc, budgetSvc))
		r.Mount("/activity", routes.ActivityRoutes(activitySvc))
		r.Mount("/dashboard", routes.DashboardRoutes(dashboardSvc))
		r.Get("/sidebar-badges", routes.SidebarBadgesHandler(dashboardSvc))
		r.Mount("/secrets", routes.SecretRoutes(secretSvc))
		r.Mount("/assets", routes.AssetRoutes(assetSvc, storageSvc))
		r.Mount("/board-settings", routes.BoardSettingsRoutes(a.DB))

		// WebSocket endpoint.
		r.Get("/events/ws", realtime.WebSocketHandler(a.Hub, a.PM))
	})
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if a.DB != nil {
		if err := db.HealthCheck(r.Context(), a.DB); err != nil {
			slog.Error("db health check failed", "err", err)
			resp := map[string]interface{}{"ok": true, "db": "disconnected"}
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]interface{}{"ok": true, "db": "connected"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := map[string]interface{}{"ok": true}
	json.NewEncoder(w).Encode(resp)
}
