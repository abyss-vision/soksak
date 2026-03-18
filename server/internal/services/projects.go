package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/domain"
)

// ProjectService handles CRUD and workspace operations for projects.
type ProjectService struct {
	db *sqlx.DB
}

// NewProjectService creates a new ProjectService.
func NewProjectService(db *sqlx.DB) *ProjectService {
	return &ProjectService{db: db}
}

// List returns all active projects for a company.
func (s *ProjectService) List(ctx context.Context, companyUUID string) ([]domain.Project, error) {
	var projects []domain.Project
	err := s.db.SelectContext(ctx, &projects,
		`SELECT * FROM projects WHERE company_uuid = $1 AND archived_at IS NULL ORDER BY created_at DESC`,
		companyUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("projects.List: %w", err)
	}
	return projects, nil
}

// Get returns a single project by UUID within a company.
func (s *ProjectService) Get(ctx context.Context, companyUUID, projectUUID string) (*domain.Project, error) {
	var project domain.Project
	err := s.db.GetContext(ctx, &project,
		`SELECT * FROM projects WHERE uuid = $1 AND company_uuid = $2`,
		projectUUID, companyUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("projects.Get: %w", err)
	}
	return &project, nil
}

// CreateProjectInput holds fields for creating a new project.
type CreateProjectInput struct {
	Name                     string          `json:"name"`
	Description              *string         `json:"description"`
	Status                   string          `json:"status"`
	GoalUUID                 *string         `json:"goalUuid"`
	LeadAgentUUID            *string         `json:"leadAgentUuid"`
	TargetDate               *string         `json:"targetDate"`
	Color                    *string         `json:"color"`
	ExecutionWorkspacePolicy json.RawMessage `json:"executionWorkspacePolicy"`
}

// Create creates a new project for a company.
func (s *ProjectService) Create(ctx context.Context, companyUUID string, input CreateProjectInput) (*domain.Project, error) {
	id := uuid.NewString()
	status := input.Status
	if status == "" {
		status = "active"
	}
	policy := input.ExecutionWorkspacePolicy
	if len(policy) == 0 {
		policy = json.RawMessage("{}")
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO projects (
			uuid, company_uuid, goal_uuid, name, description, status,
			lead_agent_uuid, target_date, color, execution_workspace_policy
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10
		)`,
		id, companyUUID, input.GoalUUID, input.Name, input.Description, status,
		input.LeadAgentUUID, input.TargetDate, input.Color, policy,
	)
	if err != nil {
		return nil, fmt.Errorf("projects.Create: %w", err)
	}
	return s.Get(ctx, companyUUID, id)
}

// UpdateProjectInput holds fields for updating a project.
type UpdateProjectInput struct {
	Name                     *string         `json:"name"`
	Description              *string         `json:"description"`
	Status                   *string         `json:"status"`
	GoalUUID                 *string         `json:"goalUuid"`
	LeadAgentUUID            *string         `json:"leadAgentUuid"`
	TargetDate               *string         `json:"targetDate"`
	Color                    *string         `json:"color"`
	PauseReason              *string         `json:"pauseReason"`
	ExecutionWorkspacePolicy json.RawMessage `json:"executionWorkspacePolicy"`
}

// Update updates a project's fields.
func (s *ProjectService) Update(ctx context.Context, companyUUID, projectUUID string, input UpdateProjectInput) (*domain.Project, error) {
	existing, err := s.Get(ctx, companyUUID, projectUUID)
	if err != nil {
		return nil, err
	}

	name := existing.Name
	if input.Name != nil {
		name = *input.Name
	}
	description := existing.Description
	if input.Description != nil {
		description = input.Description
	}
	status := existing.Status
	if input.Status != nil {
		status = *input.Status
	}
	goalUUID := existing.GoalUUID
	if input.GoalUUID != nil {
		goalUUID = input.GoalUUID
	}
	leadAgentUUID := existing.LeadAgentUUID
	if input.LeadAgentUUID != nil {
		leadAgentUUID = input.LeadAgentUUID
	}
	targetDate := existing.TargetDate
	if input.TargetDate != nil {
		targetDate = input.TargetDate
	}
	color := existing.Color
	if input.Color != nil {
		color = input.Color
	}
	pauseReason := existing.PauseReason
	if input.PauseReason != nil {
		pauseReason = input.PauseReason
	}
	policy := existing.ExecutionWorkspacePolicy
	if len(input.ExecutionWorkspacePolicy) > 0 {
		policy = input.ExecutionWorkspacePolicy
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE projects SET
			name = $1, description = $2, status = $3,
			goal_uuid = $4, lead_agent_uuid = $5,
			target_date = $6, color = $7, pause_reason = $8,
			execution_workspace_policy = $9, updated_at = now()
		WHERE uuid = $10 AND company_uuid = $11`,
		name, description, status,
		goalUUID, leadAgentUUID,
		targetDate, color, pauseReason,
		policy, projectUUID, companyUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("projects.Update: %w", err)
	}
	return s.Get(ctx, companyUUID, projectUUID)
}

// Delete removes a project by UUID.
func (s *ProjectService) Delete(ctx context.Context, companyUUID, projectUUID string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM projects WHERE uuid = $1 AND company_uuid = $2`,
		projectUUID, companyUUID,
	)
	if err != nil {
		return fmt.Errorf("projects.Delete: %w", err)
	}
	return nil
}

// ListWorkspaces returns all workspaces for a project.
func (s *ProjectService) ListWorkspaces(ctx context.Context, companyUUID, projectUUID string) ([]domain.ProjectWorkspace, error) {
	var workspaces []domain.ProjectWorkspace
	err := s.db.SelectContext(ctx, &workspaces,
		`SELECT * FROM project_workspaces WHERE project_uuid = $1 AND company_uuid = $2 ORDER BY is_primary DESC, created_at ASC`,
		projectUUID, companyUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("projects.ListWorkspaces: %w", err)
	}
	return workspaces, nil
}
