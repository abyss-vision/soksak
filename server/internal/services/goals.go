package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/domain"
	"soksak/internal/middleware"
)

// GoalService handles CRUD operations for goals.
type GoalService struct {
	db *sqlx.DB
}

// NewGoalService creates a new GoalService.
func NewGoalService(db *sqlx.DB) *GoalService {
	return &GoalService{db: db}
}

// CreateGoalInput contains the fields required to create a goal.
type CreateGoalInput struct {
	Title          string  `json:"title"          validate:"required,min=1,max=500"`
	Description    *string `json:"description"`
	Level          string  `json:"level"`
	Status         string  `json:"status"`
	ParentUUID     *string `json:"parentUuid"`
	OwnerAgentUUID *string `json:"ownerAgentUuid"`
}

// UpdateGoalInput contains the fields that can be updated on a goal.
type UpdateGoalInput struct {
	Title          *string `json:"title"          validate:"omitempty,min=1,max=500"`
	Description    *string `json:"description"`
	Level          *string `json:"level"`
	Status         *string `json:"status"`
	ParentUUID     *string `json:"parentUuid"`
	OwnerAgentUUID *string `json:"ownerAgentUuid"`
}

// List returns all goals for a company.
func (s *GoalService) List(ctx context.Context, companyUUID string) ([]domain.Goal, error) {
	var goals []domain.Goal
	err := s.db.SelectContext(ctx, &goals, `
		SELECT uuid, company_uuid, title, description, level, status,
		       parent_uuid, owner_agent_uuid, created_at, updated_at
		FROM goals
		WHERE company_uuid = $1
		ORDER BY created_at DESC
	`, companyUUID)
	if err != nil {
		return nil, err
	}
	return goals, nil
}

// Get returns a single goal by UUID within a company.
func (s *GoalService) Get(ctx context.Context, companyUUID, goalUUID string) (*domain.Goal, error) {
	var goal domain.Goal
	err := s.db.GetContext(ctx, &goal, `
		SELECT uuid, company_uuid, title, description, level, status,
		       parent_uuid, owner_agent_uuid, created_at, updated_at
		FROM goals
		WHERE uuid         = $1
		  AND company_uuid = $2
	`, goalUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Goal not found")
	}
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

// Create creates a new goal within a company.
func (s *GoalService) Create(ctx context.Context, companyUUID string, input CreateGoalInput) (*domain.Goal, error) {
	id := uuid.New().String()

	level := input.Level
	if level == "" {
		level = "task"
	}
	status := input.Status
	if status == "" {
		status = "planned"
	}

	var goal domain.Goal
	err := s.db.GetContext(ctx, &goal, `
		INSERT INTO goals (
			uuid, company_uuid, title, description, level, status,
			parent_uuid, owner_agent_uuid
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8
		)
		RETURNING uuid, company_uuid, title, description, level, status,
		          parent_uuid, owner_agent_uuid, created_at, updated_at
	`,
		id, companyUUID, input.Title, input.Description, level, status,
		input.ParentUUID, input.OwnerAgentUUID,
	)
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

// Update applies partial updates to a goal.
func (s *GoalService) Update(ctx context.Context, companyUUID, goalUUID string, input UpdateGoalInput) (*domain.Goal, error) {
	existing, err := s.Get(ctx, companyUUID, goalUUID)
	if err != nil {
		return nil, err
	}

	title := existing.Title
	if input.Title != nil {
		title = *input.Title
	}
	description := existing.Description
	if input.Description != nil {
		description = input.Description
	}
	level := existing.Level
	if input.Level != nil {
		level = *input.Level
	}
	status := existing.Status
	if input.Status != nil {
		status = *input.Status
	}
	parentUUID := existing.ParentUUID
	if input.ParentUUID != nil {
		parentUUID = input.ParentUUID
	}
	ownerAgentUUID := existing.OwnerAgentUUID
	if input.OwnerAgentUUID != nil {
		ownerAgentUUID = input.OwnerAgentUUID
	}

	var goal domain.Goal
	err = s.db.GetContext(ctx, &goal, `
		UPDATE goals
		SET title           = $1,
		    description     = $2,
		    level           = $3,
		    status          = $4,
		    parent_uuid     = $5,
		    owner_agent_uuid = $6,
		    updated_at      = now()
		WHERE uuid         = $7
		  AND company_uuid = $8
		RETURNING uuid, company_uuid, title, description, level, status,
		          parent_uuid, owner_agent_uuid, created_at, updated_at
	`,
		title, description, level, status, parentUUID, ownerAgentUUID,
		goalUUID, companyUUID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Goal not found")
	}
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

// Delete removes a goal.
func (s *GoalService) Delete(ctx context.Context, companyUUID, goalUUID string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM goals
		WHERE uuid         = $1
		  AND company_uuid = $2
	`, goalUUID, companyUUID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return middleware.ErrNotFound("Goal not found")
	}
	return nil
}

// ListLinkedIssues returns all issues linked to the given goal.
func (s *GoalService) ListLinkedIssues(ctx context.Context, companyUUID, goalUUID string) ([]domain.Issue, error) {
	// Verify goal exists in company first.
	if _, err := s.Get(ctx, companyUUID, goalUUID); err != nil {
		return nil, err
	}

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
		WHERE goal_uuid    = $1
		  AND company_uuid = $2
		ORDER BY created_at DESC
	`, goalUUID, companyUUID)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

// GoalProgress holds aggregated issue counts for a goal.
type GoalProgress struct {
	Total     int `json:"total"`
	Done      int `json:"done"`
	InProgress int `json:"inProgress"`
}

// CalculateProgress returns issue counts for the given goal.
func (s *GoalService) CalculateProgress(ctx context.Context, companyUUID, goalUUID string) (*GoalProgress, error) {
	if _, err := s.Get(ctx, companyUUID, goalUUID); err != nil {
		return nil, err
	}

	var progress GoalProgress
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*)                                          AS total,
			COUNT(*) FILTER (WHERE status = 'done')          AS done,
			COUNT(*) FILTER (WHERE status = 'in_progress')   AS in_progress
		FROM issues
		WHERE goal_uuid    = $1
		  AND company_uuid = $2
	`, goalUUID, companyUUID).Scan(&progress.Total, &progress.Done, &progress.InProgress)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}
