package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/domain"
	"soksak/internal/middleware"
)

// BudgetService handles budget policy CRUD and incident enforcement.
type BudgetService struct {
	db *sqlx.DB
}

// NewBudgetService creates a new BudgetService.
func NewBudgetService(db *sqlx.DB) *BudgetService {
	return &BudgetService{db: db}
}

// CreateBudgetPolicyInput contains the fields required to create a budget policy.
type CreateBudgetPolicyInput struct {
	ScopeType       string `json:"scopeType"       validate:"required"`
	ScopeID         string `json:"scopeId"         validate:"required"`
	Metric          string `json:"metric"          validate:"required"`
	WindowKind      string `json:"windowKind"      validate:"required"`
	Amount          int    `json:"amount"          validate:"min=1"`
	WarnPercent     int    `json:"warnPercent"     validate:"min=0,max=100"`
	HardStopEnabled bool   `json:"hardStopEnabled"`
	NotifyEnabled   bool   `json:"notifyEnabled"`
}

// UpdateBudgetPolicyInput contains fields that can be updated on a budget policy.
type UpdateBudgetPolicyInput struct {
	Amount          *int   `json:"amount"          validate:"omitempty,min=1"`
	WarnPercent     *int   `json:"warnPercent"     validate:"omitempty,min=0,max=100"`
	HardStopEnabled *bool  `json:"hardStopEnabled"`
	NotifyEnabled   *bool  `json:"notifyEnabled"`
	IsActive        *bool  `json:"isActive"`
}

// ListBudgetPolicies returns all budget policies for a company.
func (s *BudgetService) ListBudgetPolicies(ctx context.Context, companyUUID string) ([]domain.BudgetPolicy, error) {
	var policies []domain.BudgetPolicy
	err := s.db.SelectContext(ctx, &policies, `
		SELECT uuid, company_uuid, scope_type, scope_id, metric, window_kind,
		       amount, warn_percent, hard_stop_enabled, notify_enabled, is_active,
		       created_by_user_id, updated_by_user_id, created_at, updated_at
		FROM budget_policies
		WHERE company_uuid = $1
		ORDER BY created_at DESC
	`, companyUUID)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

// GetBudgetPolicy returns a single budget policy by UUID.
func (s *BudgetService) GetBudgetPolicy(ctx context.Context, companyUUID, policyUUID string) (*domain.BudgetPolicy, error) {
	var policy domain.BudgetPolicy
	err := s.db.GetContext(ctx, &policy, `
		SELECT uuid, company_uuid, scope_type, scope_id, metric, window_kind,
		       amount, warn_percent, hard_stop_enabled, notify_enabled, is_active,
		       created_by_user_id, updated_by_user_id, created_at, updated_at
		FROM budget_policies
		WHERE uuid = $1 AND company_uuid = $2
	`, policyUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Budget policy not found")
	}
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// CreateBudgetPolicy creates a new budget policy for a company.
func (s *BudgetService) CreateBudgetPolicy(ctx context.Context, companyUUID string, input CreateBudgetPolicyInput, userID *string) (*domain.BudgetPolicy, error) {
	id := uuid.New().String()
	var policy domain.BudgetPolicy
	err := s.db.GetContext(ctx, &policy, `
		INSERT INTO budget_policies (
			uuid, company_uuid, scope_type, scope_id, metric, window_kind,
			amount, warn_percent, hard_stop_enabled, notify_enabled, is_active,
			created_by_user_id
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, true,
			$11
		)
		RETURNING uuid, company_uuid, scope_type, scope_id, metric, window_kind,
		          amount, warn_percent, hard_stop_enabled, notify_enabled, is_active,
		          created_by_user_id, updated_by_user_id, created_at, updated_at
	`,
		id, companyUUID, input.ScopeType, input.ScopeID, input.Metric, input.WindowKind,
		input.Amount, input.WarnPercent, input.HardStopEnabled, input.NotifyEnabled,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// UpdateBudgetPolicy applies partial updates to a budget policy.
func (s *BudgetService) UpdateBudgetPolicy(ctx context.Context, companyUUID, policyUUID string, input UpdateBudgetPolicyInput, userID *string) (*domain.BudgetPolicy, error) {
	existing, err := s.GetBudgetPolicy(ctx, companyUUID, policyUUID)
	if err != nil {
		return nil, err
	}

	amount := existing.Amount
	if input.Amount != nil {
		amount = *input.Amount
	}
	warnPercent := existing.WarnPercent
	if input.WarnPercent != nil {
		warnPercent = *input.WarnPercent
	}
	hardStop := existing.HardStopEnabled
	if input.HardStopEnabled != nil {
		hardStop = *input.HardStopEnabled
	}
	notify := existing.NotifyEnabled
	if input.NotifyEnabled != nil {
		notify = *input.NotifyEnabled
	}
	isActive := existing.IsActive
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	var policy domain.BudgetPolicy
	err = s.db.GetContext(ctx, &policy, `
		UPDATE budget_policies
		SET amount            = $1,
		    warn_percent      = $2,
		    hard_stop_enabled = $3,
		    notify_enabled    = $4,
		    is_active         = $5,
		    updated_by_user_id = $6,
		    updated_at        = now()
		WHERE uuid = $7 AND company_uuid = $8
		RETURNING uuid, company_uuid, scope_type, scope_id, metric, window_kind,
		          amount, warn_percent, hard_stop_enabled, notify_enabled, is_active,
		          created_by_user_id, updated_by_user_id, created_at, updated_at
	`, amount, warnPercent, hardStop, notify, isActive, userID, policyUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Budget policy not found")
	}
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// DeleteBudgetPolicy removes a budget policy.
func (s *BudgetService) DeleteBudgetPolicy(ctx context.Context, companyUUID, policyUUID string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM budget_policies
		WHERE uuid = $1 AND company_uuid = $2
	`, policyUUID, companyUUID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return middleware.ErrNotFound("Budget policy not found")
	}
	return nil
}

// EnforcePolicyInput describes a threshold crossing.
type EnforcePolicyInput struct {
	PolicyUUID     string    `json:"policyUuid"     validate:"required,uuid"`
	ScopeType      string    `json:"scopeType"      validate:"required"`
	ScopeID        string    `json:"scopeId"        validate:"required"`
	Metric         string    `json:"metric"         validate:"required"`
	WindowKind     string    `json:"windowKind"     validate:"required"`
	WindowStart    time.Time `json:"windowStart"`
	WindowEnd      time.Time `json:"windowEnd"`
	ThresholdType  string    `json:"thresholdType"  validate:"required"`
	AmountLimit    int       `json:"amountLimit"    validate:"min=0"`
	AmountObserved int       `json:"amountObserved" validate:"min=0"`
}

// EnforcePolicy creates a budget incident when a threshold is crossed.
func (s *BudgetService) EnforcePolicy(ctx context.Context, companyUUID string, input EnforcePolicyInput) (*domain.BudgetIncident, error) {
	id := uuid.New().String()
	var incident domain.BudgetIncident
	err := s.db.GetContext(ctx, &incident, `
		INSERT INTO budget_incidents (
			uuid, company_uuid, policy_uuid, scope_type, scope_id, metric, window_kind,
			window_start, window_end, threshold_type, amount_limit, amount_observed, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, 'active'
		)
		RETURNING uuid, company_uuid, policy_uuid, scope_type, scope_id, metric, window_kind,
		          window_start, window_end, threshold_type, amount_limit, amount_observed,
		          status, approval_uuid, resolved_at, created_at, updated_at
	`,
		id, companyUUID, input.PolicyUUID, input.ScopeType, input.ScopeID,
		input.Metric, input.WindowKind, input.WindowStart, input.WindowEnd,
		input.ThresholdType, input.AmountLimit, input.AmountObserved,
	)
	if err != nil {
		return nil, err
	}
	return &incident, nil
}

// ListBudgetIncidents returns active incidents for a company.
func (s *BudgetService) ListBudgetIncidents(ctx context.Context, companyUUID string) ([]domain.BudgetIncident, error) {
	var incidents []domain.BudgetIncident
	err := s.db.SelectContext(ctx, &incidents, `
		SELECT uuid, company_uuid, policy_uuid, scope_type, scope_id, metric, window_kind,
		       window_start, window_end, threshold_type, amount_limit, amount_observed,
		       status, approval_uuid, resolved_at, created_at, updated_at
		FROM budget_incidents
		WHERE company_uuid = $1
		ORDER BY created_at DESC
	`, companyUUID)
	if err != nil {
		return nil, err
	}
	return incidents, nil
}
