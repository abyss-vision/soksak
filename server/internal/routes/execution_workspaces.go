package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"abyss-view/internal/domain"
	"abyss-view/internal/middleware"
)

// ExecutionWorkspaceRoutes returns a chi.Router for execution workspace CRUD.
// Must be mounted under /api/companies/{companyUuid}/execution-workspaces.
func ExecutionWorkspaceRoutes(db *sqlx.DB) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := middleware.CompanyFromContext(r.Context())

		q := r.URL.Query()
		workspaces, err := listExecutionWorkspaces(r.Context(), db, companyUUID, listWorkspacesFilter{
			ProjectUUID:          q.Get("projectUuid"),
			ProjectWorkspaceUUID: q.Get("projectWorkspaceUuid"),
			SourceIssueUUID:      q.Get("sourceIssueUuid"),
			Status:               q.Get("status"),
		})
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to list execution workspaces"))
			return
		}
		writeJSON(w, http.StatusOK, workspaces)
	})

	r.Get("/{workspaceUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := middleware.CompanyFromContext(r.Context())
		workspaceUUID := chi.URLParam(r, "workspaceUuid")

		workspace, err := getExecutionWorkspace(r.Context(), db, companyUUID, workspaceUUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.WriteAppError(w, r, middleware.ErrNotFound("Execution workspace not found"))
				return
			}
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get execution workspace"))
			return
		}
		writeJSON(w, http.StatusOK, workspace)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := middleware.CompanyFromContext(r.Context())

		var input createWorkspaceInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}
		if input.ProjectUUID == "" || input.Name == "" || input.Mode == "" || input.StrategyType == "" {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("projectUuid, name, mode, and strategyType are required"))
			return
		}

		workspace, err := createExecutionWorkspace(r.Context(), db, companyUUID, input)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to create execution workspace"))
			return
		}
		writeJSON(w, http.StatusCreated, workspace)
	})

	r.Patch("/{workspaceUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := middleware.CompanyFromContext(r.Context())
		workspaceUUID := chi.URLParam(r, "workspaceUuid")

		var patch updateWorkspaceInput
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		workspace, err := updateExecutionWorkspace(r.Context(), db, companyUUID, workspaceUUID, patch)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.WriteAppError(w, r, middleware.ErrNotFound("Execution workspace not found"))
				return
			}
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update execution workspace"))
			return
		}
		writeJSON(w, http.StatusOK, workspace)
	})

	r.Delete("/{workspaceUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := middleware.CompanyFromContext(r.Context())
		workspaceUUID := chi.URLParam(r, "workspaceUuid")

		result, err := db.ExecContext(r.Context(), `
			UPDATE execution_workspaces
			SET status     = 'archived',
			    closed_at  = now(),
			    updated_at = now()
			WHERE uuid         = $1
			  AND company_uuid = $2
			  AND status      != 'archived'
		`, workspaceUUID, companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to archive execution workspace"))
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			middleware.WriteAppError(w, r, middleware.ErrNotFound("Execution workspace not found"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}

type listWorkspacesFilter struct {
	ProjectUUID          string
	ProjectWorkspaceUUID string
	SourceIssueUUID      string
	Status               string
}

type createWorkspaceInput struct {
	ProjectUUID          string          `json:"projectUuid"`
	ProjectWorkspaceUUID *string         `json:"projectWorkspaceUuid"`
	SourceIssueUUID      *string         `json:"sourceIssueUuid"`
	Mode                 string          `json:"mode"`
	StrategyType         string          `json:"strategyType"`
	Name                 string          `json:"name"`
	Cwd                  *string         `json:"cwd"`
	RepoURL              *string         `json:"repoUrl"`
	BaseRef              *string         `json:"baseRef"`
	BranchName           *string         `json:"branchName"`
	ProviderType         string          `json:"providerType"`
	ProviderRef          *string         `json:"providerRef"`
	Metadata             json.RawMessage `json:"metadata"`
}

type updateWorkspaceInput struct {
	Status              *string         `json:"status"`
	Cwd                 *string         `json:"cwd"`
	RepoURL             *string         `json:"repoUrl"`
	BaseRef             *string         `json:"baseRef"`
	BranchName          *string         `json:"branchName"`
	ProviderRef         *string         `json:"providerRef"`
	CleanupEligibleAt   *time.Time      `json:"cleanupEligibleAt"`
	CleanupReason       *string         `json:"cleanupReason"`
	Metadata            json.RawMessage `json:"metadata"`
}

func listExecutionWorkspaces(ctx context.Context, db *sqlx.DB, companyUUID string, f listWorkspacesFilter) ([]domain.ExecutionWorkspace, error) {
	query := `
		SELECT uuid, company_uuid, project_uuid, project_workspace_uuid, source_issue_uuid,
		       mode, strategy_type, name, status, cwd, repo_url, base_ref, branch_name,
		       provider_type, provider_ref, derived_from_execution_workspace_uuid,
		       last_used_at, opened_at, closed_at, cleanup_eligible_at, cleanup_reason,
		       metadata, created_at, updated_at
		FROM execution_workspaces
		WHERE company_uuid = $1
	`
	args := []interface{}{companyUUID}
	idx := 2

	if f.ProjectUUID != "" {
		query += ` AND project_uuid = $` + itoa(idx)
		args = append(args, f.ProjectUUID)
		idx++
	}
	if f.ProjectWorkspaceUUID != "" {
		query += ` AND project_workspace_uuid = $` + itoa(idx)
		args = append(args, f.ProjectWorkspaceUUID)
		idx++
	}
	if f.SourceIssueUUID != "" {
		query += ` AND source_issue_uuid = $` + itoa(idx)
		args = append(args, f.SourceIssueUUID)
		idx++
	}
	if f.Status != "" {
		query += ` AND status = $` + itoa(idx)
		args = append(args, f.Status)
		idx++
	}
	_ = idx

	query += ` ORDER BY created_at DESC`

	var workspaces []domain.ExecutionWorkspace
	if err := db.SelectContext(ctx, &workspaces, query, args...); err != nil {
		return nil, err
	}
	return workspaces, nil
}

func getExecutionWorkspace(ctx context.Context, db *sqlx.DB, companyUUID, workspaceUUID string) (*domain.ExecutionWorkspace, error) {
	var workspace domain.ExecutionWorkspace
	err := db.GetContext(ctx, &workspace, `
		SELECT uuid, company_uuid, project_uuid, project_workspace_uuid, source_issue_uuid,
		       mode, strategy_type, name, status, cwd, repo_url, base_ref, branch_name,
		       provider_type, provider_ref, derived_from_execution_workspace_uuid,
		       last_used_at, opened_at, closed_at, cleanup_eligible_at, cleanup_reason,
		       metadata, created_at, updated_at
		FROM execution_workspaces
		WHERE uuid         = $1
		  AND company_uuid = $2
	`, workspaceUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Execution workspace not found")
	}
	return &workspace, err
}

func createExecutionWorkspace(ctx context.Context, db *sqlx.DB, companyUUID string, input createWorkspaceInput) (*domain.ExecutionWorkspace, error) {
	id := uuid.New().String()
	providerType := input.ProviderType
	if providerType == "" {
		providerType = "local_fs"
	}
	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}

	var workspace domain.ExecutionWorkspace
	err := db.GetContext(ctx, &workspace, `
		INSERT INTO execution_workspaces
		    (uuid, company_uuid, project_uuid, project_workspace_uuid, source_issue_uuid,
		     mode, strategy_type, name, status, cwd, repo_url, base_ref, branch_name,
		     provider_type, provider_ref, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'active',$9,$10,$11,$12,$13,$14,$15)
		RETURNING uuid, company_uuid, project_uuid, project_workspace_uuid, source_issue_uuid,
		          mode, strategy_type, name, status, cwd, repo_url, base_ref, branch_name,
		          provider_type, provider_ref, derived_from_execution_workspace_uuid,
		          last_used_at, opened_at, closed_at, cleanup_eligible_at, cleanup_reason,
		          metadata, created_at, updated_at
	`,
		id, companyUUID, input.ProjectUUID, input.ProjectWorkspaceUUID, input.SourceIssueUUID,
		input.Mode, input.StrategyType, input.Name,
		input.Cwd, input.RepoURL, input.BaseRef, input.BranchName,
		providerType, input.ProviderRef, metadata,
	)
	return &workspace, err
}

func updateExecutionWorkspace(ctx context.Context, db *sqlx.DB, companyUUID, workspaceUUID string, patch updateWorkspaceInput) (*domain.ExecutionWorkspace, error) {
	existing, err := getExecutionWorkspace(ctx, db, companyUUID, workspaceUUID)
	if err != nil {
		return nil, err
	}

	status := existing.Status
	if patch.Status != nil {
		status = *patch.Status
	}
	cwd := existing.Cwd
	if patch.Cwd != nil {
		cwd = patch.Cwd
	}
	repoURL := existing.RepoURL
	if patch.RepoURL != nil {
		repoURL = patch.RepoURL
	}
	baseRef := existing.BaseRef
	if patch.BaseRef != nil {
		baseRef = patch.BaseRef
	}
	branchName := existing.BranchName
	if patch.BranchName != nil {
		branchName = patch.BranchName
	}
	providerRef := existing.ProviderRef
	if patch.ProviderRef != nil {
		providerRef = patch.ProviderRef
	}
	cleanupEligibleAt := existing.CleanupEligibleAt
	if patch.CleanupEligibleAt != nil {
		cleanupEligibleAt = patch.CleanupEligibleAt
	}
	cleanupReason := existing.CleanupReason
	if patch.CleanupReason != nil {
		cleanupReason = patch.CleanupReason
	}
	metadata := existing.Metadata
	if len(patch.Metadata) > 0 {
		metadata = patch.Metadata
	}

	var closedAt *time.Time
	if existing.ClosedAt != nil {
		closedAt = existing.ClosedAt
	}
	if status == "archived" && existing.Status != "archived" {
		now := time.Now()
		closedAt = &now
	}

	var workspace domain.ExecutionWorkspace
	err = db.GetContext(ctx, &workspace, `
		UPDATE execution_workspaces
		SET status               = $1,
		    cwd                  = $2,
		    repo_url             = $3,
		    base_ref             = $4,
		    branch_name          = $5,
		    provider_ref         = $6,
		    cleanup_eligible_at  = $7,
		    cleanup_reason       = $8,
		    metadata             = $9,
		    closed_at            = $10,
		    updated_at           = now()
		WHERE uuid         = $11
		  AND company_uuid = $12
		RETURNING uuid, company_uuid, project_uuid, project_workspace_uuid, source_issue_uuid,
		          mode, strategy_type, name, status, cwd, repo_url, base_ref, branch_name,
		          provider_type, provider_ref, derived_from_execution_workspace_uuid,
		          last_used_at, opened_at, closed_at, cleanup_eligible_at, cleanup_reason,
		          metadata, created_at, updated_at
	`,
		status, cwd, repoURL, baseRef, branchName, providerRef,
		cleanupEligibleAt, cleanupReason, metadata, closedAt,
		workspaceUUID, companyUUID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Execution workspace not found")
	}
	return &workspace, err
}

// itoa converts an int to its decimal string representation without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
