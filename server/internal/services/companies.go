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

// CompanyService handles CRUD operations for companies.
type CompanyService struct {
	db *sqlx.DB
}

// NewCompanyService creates a new CompanyService.
func NewCompanyService(db *sqlx.DB) *CompanyService {
	return &CompanyService{db: db}
}

// CreateCompanyInput contains the fields required to create a company.
type CreateCompanyInput struct {
	Name                             string  `json:"name"                             validate:"required,min=1,max=255"`
	Description                      *string `json:"description"`
	IssuePrefix                      string  `json:"issuePrefix"                      validate:"required,min=1,max=10"`
	BudgetMonthlyCents               int     `json:"budgetMonthlyCents"`
	RequireBoardApprovalForNewAgents bool    `json:"requireBoardApprovalForNewAgents"`
	BrandColor                       *string `json:"brandColor"`
}

// UpdateCompanyInput contains the fields that can be updated on a company.
type UpdateCompanyInput struct {
	Name                             *string `json:"name"                             validate:"omitempty,min=1,max=255"`
	Description                      *string `json:"description"`
	Status                           *string `json:"status"`
	PauseReason                      *string `json:"pauseReason"`
	BudgetMonthlyCents               *int    `json:"budgetMonthlyCents"`
	RequireBoardApprovalForNewAgents *bool   `json:"requireBoardApprovalForNewAgents"`
	BrandColor                       *string `json:"brandColor"`
}

// List returns all non-deleted companies.
func (s *CompanyService) List(ctx context.Context) ([]domain.Company, error) {
	var companies []domain.Company
	err := s.db.SelectContext(ctx, &companies, `
		SELECT uuid, name, description, status, pause_reason, paused_at,
		       issue_prefix, issue_counter, budget_monthly_cents, spent_monthly_cents,
		       require_board_approval_for_new_agents, brand_color, created_at, updated_at
		FROM companies
		WHERE status != 'deleted'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	return companies, nil
}

// Get returns a single company by UUID.
func (s *CompanyService) Get(ctx context.Context, companyUUID string) (*domain.Company, error) {
	var company domain.Company
	err := s.db.GetContext(ctx, &company, `
		SELECT uuid, name, description, status, pause_reason, paused_at,
		       issue_prefix, issue_counter, budget_monthly_cents, spent_monthly_cents,
		       require_board_approval_for_new_agents, brand_color, created_at, updated_at
		FROM companies
		WHERE uuid = $1
		  AND status != 'deleted'
	`, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Company not found")
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// Create creates a new company with a generated UUID v4.
func (s *CompanyService) Create(ctx context.Context, input CreateCompanyInput) (*domain.Company, error) {
	id := uuid.New().String()
	var company domain.Company
	err := s.db.GetContext(ctx, &company, `
		INSERT INTO companies (
			uuid, name, description, status, issue_prefix,
			budget_monthly_cents, require_board_approval_for_new_agents, brand_color
		) VALUES (
			$1, $2, $3, 'active', $4,
			$5, $6, $7
		)
		RETURNING uuid, name, description, status, pause_reason, paused_at,
		          issue_prefix, issue_counter, budget_monthly_cents, spent_monthly_cents,
		          require_board_approval_for_new_agents, brand_color, created_at, updated_at
	`,
		id,
		input.Name,
		input.Description,
		input.IssuePrefix,
		input.BudgetMonthlyCents,
		input.RequireBoardApprovalForNewAgents,
		input.BrandColor,
	)
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// Update applies partial updates to a company.
func (s *CompanyService) Update(ctx context.Context, companyUUID string, input UpdateCompanyInput) (*domain.Company, error) {
	existing, err := s.Get(ctx, companyUUID)
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
	pauseReason := existing.PauseReason
	if input.PauseReason != nil {
		pauseReason = input.PauseReason
	}
	budgetMonthlyCents := existing.BudgetMonthlyCents
	if input.BudgetMonthlyCents != nil {
		budgetMonthlyCents = *input.BudgetMonthlyCents
	}
	requireApproval := existing.RequireBoardApprovalForNewAgents
	if input.RequireBoardApprovalForNewAgents != nil {
		requireApproval = *input.RequireBoardApprovalForNewAgents
	}
	brandColor := existing.BrandColor
	if input.BrandColor != nil {
		brandColor = input.BrandColor
	}

	var company domain.Company
	err = s.db.GetContext(ctx, &company, `
		UPDATE companies
		SET name                               = $1,
		    description                        = $2,
		    status                             = $3,
		    pause_reason                       = $4,
		    budget_monthly_cents               = $5,
		    require_board_approval_for_new_agents = $6,
		    brand_color                        = $7,
		    updated_at                         = now()
		WHERE uuid = $8
		  AND status != 'deleted'
		RETURNING uuid, name, description, status, pause_reason, paused_at,
		          issue_prefix, issue_counter, budget_monthly_cents, spent_monthly_cents,
		          require_board_approval_for_new_agents, brand_color, created_at, updated_at
	`,
		name, description, status, pauseReason,
		budgetMonthlyCents, requireApproval, brandColor,
		companyUUID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Company not found")
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// Delete soft-deletes a company by setting status to 'deleted'.
func (s *CompanyService) Delete(ctx context.Context, companyUUID string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE companies
		SET status     = 'deleted',
		    updated_at = now()
		WHERE uuid   = $1
		  AND status != 'deleted'
	`, companyUUID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return middleware.ErrNotFound("Company not found")
	}
	return nil
}

// ListMembers returns the active memberships for a company.
func (s *CompanyService) ListMembers(ctx context.Context, companyUUID string) ([]domain.CompanyMembership, error) {
	var members []domain.CompanyMembership
	err := s.db.SelectContext(ctx, &members, `
		SELECT uuid, company_uuid, principal_type, principal_id,
		       status, membership_role, created_at, updated_at
		FROM company_memberships
		WHERE company_uuid = $1
		  AND status       = 'active'
		ORDER BY created_at ASC
	`, companyUUID)
	if err != nil {
		return nil, err
	}
	return members, nil
}
