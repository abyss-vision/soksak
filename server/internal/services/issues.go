package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/domain"
	"soksak/internal/middleware"
	"soksak/internal/realtime"
)

// IssueService handles CRUD operations and state transitions for issues.
type IssueService struct {
	db  *sqlx.DB
	hub *realtime.Hub
}

// NewIssueService creates a new IssueService.
func NewIssueService(db *sqlx.DB, hub *realtime.Hub) *IssueService {
	return &IssueService{db: db, hub: hub}
}

// CreateIssueInput contains the fields required to create an issue.
type CreateIssueInput struct {
	Title                       string          `json:"title"                       validate:"required,min=1,max=500"`
	Description                 *string         `json:"description"`
	Status                      string          `json:"status"`
	Priority                    string          `json:"priority"`
	ProjectUUID                 *string         `json:"projectUuid"`
	GoalUUID                    *string         `json:"goalUuid"`
	ParentUUID                  *string         `json:"parentUuid"`
	AssigneeAgentUUID           *string         `json:"assigneeAgentUuid"`
	AssigneeUserID              *string         `json:"assigneeUserId"`
	BillingCode                 *string         `json:"billingCode"`
	AssigneeAdapterOverrides    json.RawMessage `json:"assigneeAdapterOverrides"`
	ExecutionWorkspacePreference *string        `json:"executionWorkspacePreference"`
}

// UpdateIssueInput contains the fields that can be updated on an issue.
type UpdateIssueInput struct {
	Title                        *string         `json:"title"                        validate:"omitempty,min=1,max=500"`
	Description                  *string         `json:"description"`
	Status                       *string         `json:"status"`
	Priority                     *string         `json:"priority"`
	ProjectUUID                  *string         `json:"projectUuid"`
	GoalUUID                     *string         `json:"goalUuid"`
	ParentUUID                   *string         `json:"parentUuid"`
	AssigneeAgentUUID            *string         `json:"assigneeAgentUuid"`
	AssigneeUserID               *string         `json:"assigneeUserId"`
	BillingCode                  *string         `json:"billingCode"`
	AssigneeAdapterOverrides     json.RawMessage `json:"assigneeAdapterOverrides"`
	ExecutionWorkspacePreference *string         `json:"executionWorkspacePreference"`
}

// List returns all issues for a company, optionally filtered.
func (s *IssueService) List(ctx context.Context, companyUUID string) ([]domain.Issue, error) {
	var issues []domain.Issue
	err := s.db.SelectContext(ctx, &issues, `
		SELECT uuid, company_uuid, project_uuid, project_workspace_uuid, goal_uuid, parent_uuid,
		       title, description, status, priority,
		       assignee_agent_uuid, assignee_user_id,
		       checkout_run_uuid, execution_run_uuid, execution_agent_name_key,
		       execution_locked_at, created_by_agent_uuid, created_by_user_id,
		       issue_number, identifier, request_depth, billing_code,
		       assignee_adapter_overrides, execution_workspace_uuid,
		       execution_workspace_preference, execution_workspace_settings,
		       hidden_at, started_at, completed_at, cancelled_at,
		       created_at, updated_at
		FROM issues
		WHERE company_uuid = $1
		ORDER BY created_at DESC
	`, companyUUID)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

// Get returns a single issue by UUID within a company.
func (s *IssueService) Get(ctx context.Context, companyUUID, issueUUID string) (*domain.Issue, error) {
	var issue domain.Issue
	err := s.db.GetContext(ctx, &issue, `
		SELECT uuid, company_uuid, project_uuid, project_workspace_uuid, goal_uuid, parent_uuid,
		       title, description, status, priority,
		       assignee_agent_uuid, assignee_user_id,
		       checkout_run_uuid, execution_run_uuid, execution_agent_name_key,
		       execution_locked_at, created_by_agent_uuid, created_by_user_id,
		       issue_number, identifier, request_depth, billing_code,
		       assignee_adapter_overrides, execution_workspace_uuid,
		       execution_workspace_preference, execution_workspace_settings,
		       hidden_at, started_at, completed_at, cancelled_at,
		       created_at, updated_at
		FROM issues
		WHERE uuid         = $1
		  AND company_uuid = $2
	`, issueUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Issue not found")
	}
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// Create creates a new issue within a company.
func (s *IssueService) Create(ctx context.Context, companyUUID string, input CreateIssueInput) (*domain.Issue, error) {
	id := uuid.New().String()

	status := input.Status
	if status == "" {
		status = string(domain.IssueStatusBacklog)
	}
	priority := input.Priority
	if priority == "" {
		priority = string(domain.PriorityNone)
	}

	adapterOverrides := input.AssigneeAdapterOverrides
	if adapterOverrides == nil {
		adapterOverrides = json.RawMessage("{}")
	}

	var issue domain.Issue
	err := s.db.GetContext(ctx, &issue, `
		INSERT INTO issues (
			uuid, company_uuid, project_uuid, goal_uuid, parent_uuid,
			title, description, status, priority,
			assignee_agent_uuid, assignee_user_id,
			billing_code, assignee_adapter_overrides, execution_workspace_preference
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12, $13, $14
		)
		RETURNING uuid, company_uuid, project_uuid, project_workspace_uuid, goal_uuid, parent_uuid,
		          title, description, status, priority,
		          assignee_agent_uuid, assignee_user_id,
		          checkout_run_uuid, execution_run_uuid, execution_agent_name_key,
		          execution_locked_at, created_by_agent_uuid, created_by_user_id,
		          issue_number, identifier, request_depth, billing_code,
		          assignee_adapter_overrides, execution_workspace_uuid,
		          execution_workspace_preference, execution_workspace_settings,
		          hidden_at, started_at, completed_at, cancelled_at,
		          created_at, updated_at
	`,
		id, companyUUID, input.ProjectUUID, input.GoalUUID, input.ParentUUID,
		input.Title, input.Description, status, priority,
		input.AssigneeAgentUUID, input.AssigneeUserID,
		input.BillingCode, adapterOverrides, input.ExecutionWorkspacePreference,
	)
	if err != nil {
		return nil, err
	}

	s.broadcastIssueUpdated(&issue)
	return &issue, nil
}

// Update applies partial updates to an issue, validating status transitions.
func (s *IssueService) Update(ctx context.Context, companyUUID, issueUUID string, input UpdateIssueInput) (*domain.Issue, error) {
	existing, err := s.Get(ctx, companyUUID, issueUUID)
	if err != nil {
		return nil, err
	}

	// Validate status transition before applying any changes.
	if input.Status != nil && *input.Status != existing.Status {
		if err := ValidateTransition(existing.Status, *input.Status); err != nil {
			return nil, &middleware.AppError{
				Code:    "INVALID_TRANSITION",
				Status:  422,
				Message: err.Error(),
			}
		}
	}

	// Apply field updates, falling back to existing values.
	title := existing.Title
	if input.Title != nil {
		title = *input.Title
	}
	description := existing.Description
	if input.Description != nil {
		description = input.Description
	}
	status := existing.Status
	if input.Status != nil {
		status = *input.Status
	}
	priority := existing.Priority
	if input.Priority != nil {
		priority = *input.Priority
	}
	projectUUID := existing.ProjectUUID
	if input.ProjectUUID != nil {
		projectUUID = input.ProjectUUID
	}
	goalUUID := existing.GoalUUID
	if input.GoalUUID != nil {
		goalUUID = input.GoalUUID
	}
	parentUUID := existing.ParentUUID
	if input.ParentUUID != nil {
		parentUUID = input.ParentUUID
	}
	assigneeAgentUUID := existing.AssigneeAgentUUID
	if input.AssigneeAgentUUID != nil {
		assigneeAgentUUID = input.AssigneeAgentUUID
	}
	assigneeUserID := existing.AssigneeUserID
	if input.AssigneeUserID != nil {
		assigneeUserID = input.AssigneeUserID
	}
	billingCode := existing.BillingCode
	if input.BillingCode != nil {
		billingCode = input.BillingCode
	}
	adapterOverrides := existing.AssigneeAdapterOverrides
	if input.AssigneeAdapterOverrides != nil {
		adapterOverrides = input.AssigneeAdapterOverrides
	}
	workspacePref := existing.ExecutionWorkspacePreference
	if input.ExecutionWorkspacePreference != nil {
		workspacePref = input.ExecutionWorkspacePreference
	}

	// Compute status side-effects.
	now := time.Now().UTC()
	startedAt := existing.StartedAt
	completedAt := existing.CompletedAt
	cancelledAt := existing.CancelledAt

	if input.Status != nil && *input.Status != existing.Status {
		switch domain.IssueStatus(status) {
		case domain.IssueStatusInProgress:
			if startedAt == nil {
				startedAt = &now
			}
		case domain.IssueStatusDone:
			if completedAt == nil {
				completedAt = &now
			}
		case domain.IssueStatusCancelled:
			if cancelledAt == nil {
				cancelledAt = &now
			}
		}
	}

	var issue domain.Issue
	err = s.db.GetContext(ctx, &issue, `
		UPDATE issues
		SET title                         = $1,
		    description                   = $2,
		    status                        = $3,
		    priority                      = $4,
		    project_uuid                  = $5,
		    goal_uuid                     = $6,
		    parent_uuid                   = $7,
		    assignee_agent_uuid           = $8,
		    assignee_user_id              = $9,
		    billing_code                  = $10,
		    assignee_adapter_overrides    = $11,
		    execution_workspace_preference = $12,
		    started_at                    = $13,
		    completed_at                  = $14,
		    cancelled_at                  = $15,
		    updated_at                    = now()
		WHERE uuid         = $16
		  AND company_uuid = $17
		RETURNING uuid, company_uuid, project_uuid, project_workspace_uuid, goal_uuid, parent_uuid,
		          title, description, status, priority,
		          assignee_agent_uuid, assignee_user_id,
		          checkout_run_uuid, execution_run_uuid, execution_agent_name_key,
		          execution_locked_at, created_by_agent_uuid, created_by_user_id,
		          issue_number, identifier, request_depth, billing_code,
		          assignee_adapter_overrides, execution_workspace_uuid,
		          execution_workspace_preference, execution_workspace_settings,
		          hidden_at, started_at, completed_at, cancelled_at,
		          created_at, updated_at
	`,
		title, description, status, priority,
		projectUUID, goalUUID, parentUUID,
		assigneeAgentUUID, assigneeUserID,
		billingCode, adapterOverrides, workspacePref,
		startedAt, completedAt, cancelledAt,
		issueUUID, companyUUID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Issue not found")
	}
	if err != nil {
		return nil, err
	}

	s.broadcastIssueUpdated(&issue)
	return &issue, nil
}

// Delete removes an issue.
func (s *IssueService) Delete(ctx context.Context, companyUUID, issueUUID string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM issues
		WHERE uuid         = $1
		  AND company_uuid = $2
	`, issueUUID, companyUUID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return middleware.ErrNotFound("Issue not found")
	}
	return nil
}

// broadcastIssueUpdated sends an issue.updated event to all company WebSocket clients.
func (s *IssueService) broadcastIssueUpdated(issue *domain.Issue) {
	if s.hub == nil {
		return
	}
	payload, err := json.Marshal(issue)
	if err != nil {
		return
	}
	s.hub.PublishToCompany(issue.CompanyUUID, realtime.WebSocketMessage{
		Type:      realtime.EventIssueUpdated,
		CompanyID: issue.CompanyUUID,
		Payload:   json.RawMessage(payload),
	})
}
