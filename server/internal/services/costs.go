package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"abyss-view/internal/domain"
	"abyss-view/internal/middleware"
)

// CostService handles recording and querying AI inference costs.
type CostService struct {
	db *sqlx.DB
}

// NewCostService creates a new CostService.
func NewCostService(db *sqlx.DB) *CostService {
	return &CostService{db: db}
}

// RecordCostInput contains the fields required to record a cost event.
type RecordCostInput struct {
	AgentUUID         string   `json:"agentUuid"         validate:"required,uuid"`
	IssueUUID         *string  `json:"issueUuid"`
	ProjectUUID       *string  `json:"projectUuid"`
	GoalUUID          *string  `json:"goalUuid"`
	HeartbeatRunUUID  *string  `json:"heartbeatRunUuid"`
	BillingCode       *string  `json:"billingCode"`
	Provider          string   `json:"provider"          validate:"required"`
	Biller            string   `json:"biller"            validate:"required"`
	BillingType       string   `json:"billingType"       validate:"required"`
	Model             string   `json:"model"             validate:"required"`
	InputTokens       int      `json:"inputTokens"       validate:"min=0"`
	CachedInputTokens int      `json:"cachedInputTokens" validate:"min=0"`
	OutputTokens      int      `json:"outputTokens"      validate:"min=0"`
	CostCents         int      `json:"costCents"         validate:"min=0"`
	OccurredAt        time.Time `json:"occurredAt"`
}

// RecordCost inserts a new cost event for the given company.
func (s *CostService) RecordCost(ctx context.Context, companyUUID string, input RecordCostInput) (*domain.CostEvent, error) {
	id := uuid.New().String()
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	}
	var event domain.CostEvent
	err := s.db.GetContext(ctx, &event, `
		INSERT INTO cost_events (
			uuid, company_uuid, agent_uuid, issue_uuid, project_uuid, goal_uuid,
			heartbeat_run_uuid, billing_code, provider, biller, billing_type, model,
			input_tokens, cached_input_tokens, output_tokens, cost_cents, occurred_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17
		)
		RETURNING uuid, company_uuid, agent_uuid, issue_uuid, project_uuid, goal_uuid,
		          heartbeat_run_uuid, billing_code, provider, biller, billing_type, model,
		          input_tokens, cached_input_tokens, output_tokens, cost_cents,
		          occurred_at, created_at
	`,
		id, companyUUID, input.AgentUUID, input.IssueUUID, input.ProjectUUID, input.GoalUUID,
		input.HeartbeatRunUUID, input.BillingCode, input.Provider, input.Biller,
		input.BillingType, input.Model,
		input.InputTokens, input.CachedInputTokens, input.OutputTokens,
		input.CostCents, input.OccurredAt,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ListCosts returns cost events for a company with optional filters.
func (s *CostService) ListCosts(ctx context.Context, companyUUID string, limit int) ([]domain.CostEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var events []domain.CostEvent
	err := s.db.SelectContext(ctx, &events, `
		SELECT uuid, company_uuid, agent_uuid, issue_uuid, project_uuid, goal_uuid,
		       heartbeat_run_uuid, billing_code, provider, biller, billing_type, model,
		       input_tokens, cached_input_tokens, output_tokens, cost_cents,
		       occurred_at, created_at
		FROM cost_events
		WHERE company_uuid = $1
		ORDER BY occurred_at DESC
		LIMIT $2
	`, companyUUID, limit)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// CostSummary holds aggregate cost totals.
type CostSummary struct {
	TotalCents        int `db:"total_cents"         json:"totalCents"`
	TotalInputTokens  int `db:"total_input_tokens"  json:"totalInputTokens"`
	TotalOutputTokens int `db:"total_output_tokens" json:"totalOutputTokens"`
	EventCount        int `db:"event_count"         json:"eventCount"`
}

// AgentCostRow holds per-agent cost aggregation.
type AgentCostRow struct {
	AgentUUID  string `db:"agent_uuid"  json:"agentUuid"`
	TotalCents int    `db:"total_cents" json:"totalCents"`
	EventCount int    `db:"event_count" json:"eventCount"`
}

// GetCostSummary returns aggregated cost totals for a company within an optional date range.
func (s *CostService) GetCostSummary(ctx context.Context, companyUUID string, from, to *time.Time) (*CostSummary, error) {
	var summary CostSummary
	err := s.db.GetContext(ctx, &summary, `
		SELECT
			COALESCE(SUM(cost_cents), 0)::int      AS total_cents,
			COALESCE(SUM(input_tokens), 0)::int    AS total_input_tokens,
			COALESCE(SUM(output_tokens), 0)::int   AS total_output_tokens,
			COUNT(*)::int                          AS event_count
		FROM cost_events
		WHERE company_uuid = $1
		  AND ($2::timestamptz IS NULL OR occurred_at >= $2)
		  AND ($3::timestamptz IS NULL OR occurred_at <= $3)
	`, companyUUID, from, to)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// GetCostsByAgent returns per-agent cost aggregation.
func (s *CostService) GetCostsByAgent(ctx context.Context, companyUUID string, from, to *time.Time) ([]AgentCostRow, error) {
	var rows []AgentCostRow
	err := s.db.SelectContext(ctx, &rows, `
		SELECT
			agent_uuid,
			COALESCE(SUM(cost_cents), 0)::int AS total_cents,
			COUNT(*)::int                      AS event_count
		FROM cost_events
		WHERE company_uuid = $1
		  AND ($2::timestamptz IS NULL OR occurred_at >= $2)
		  AND ($3::timestamptz IS NULL OR occurred_at <= $3)
		GROUP BY agent_uuid
		ORDER BY total_cents DESC
	`, companyUUID, from, to)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GetCostEvent returns a single cost event by UUID.
func (s *CostService) GetCostEvent(ctx context.Context, companyUUID, eventUUID string) (*domain.CostEvent, error) {
	var event domain.CostEvent
	err := s.db.GetContext(ctx, &event, `
		SELECT uuid, company_uuid, agent_uuid, issue_uuid, project_uuid, goal_uuid,
		       heartbeat_run_uuid, billing_code, provider, biller, billing_type, model,
		       input_tokens, cached_input_tokens, output_tokens, cost_cents,
		       occurred_at, created_at
		FROM cost_events
		WHERE uuid = $1 AND company_uuid = $2
	`, eventUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Cost event not found")
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}
