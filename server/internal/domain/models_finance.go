package domain

import (
	"encoding/json"
	"time"
)

// BudgetPolicy defines a spending limit for a scope (company, agent, etc.).
type BudgetPolicy struct {
	UUID             string    `db:"uuid"               json:"uuid"`
	CompanyUUID      string    `db:"company_uuid"       json:"companyUuid"`
	ScopeType        string    `db:"scope_type"         json:"scopeType"`
	ScopeID          string    `db:"scope_id"           json:"scopeId"`
	Metric           string    `db:"metric"             json:"metric"`
	WindowKind       string    `db:"window_kind"        json:"windowKind"`
	Amount           int       `db:"amount"             json:"amount"`
	WarnPercent      int       `db:"warn_percent"       json:"warnPercent"`
	HardStopEnabled  bool      `db:"hard_stop_enabled"  json:"hardStopEnabled"`
	NotifyEnabled    bool      `db:"notify_enabled"     json:"notifyEnabled"`
	IsActive         bool      `db:"is_active"          json:"isActive"`
	CreatedByUserID  *string   `db:"created_by_user_id" json:"createdByUserId"`
	UpdatedByUserID  *string   `db:"updated_by_user_id" json:"updatedByUserId"`
	CreatedAt        time.Time `db:"created_at"         json:"createdAt"`
	UpdatedAt        time.Time `db:"updated_at"         json:"updatedAt"`
}

// BudgetIncident records when a budget policy threshold was crossed.
type BudgetIncident struct {
	UUID              string     `db:"uuid"               json:"uuid"`
	CompanyUUID       string     `db:"company_uuid"       json:"companyUuid"`
	PolicyUUID        string     `db:"policy_uuid"        json:"policyUuid"`
	ScopeType         string     `db:"scope_type"         json:"scopeType"`
	ScopeID           string     `db:"scope_id"           json:"scopeId"`
	Metric            string     `db:"metric"             json:"metric"`
	WindowKind        string     `db:"window_kind"        json:"windowKind"`
	WindowStart       time.Time  `db:"window_start"       json:"windowStart"`
	WindowEnd         time.Time  `db:"window_end"         json:"windowEnd"`
	ThresholdType     string     `db:"threshold_type"     json:"thresholdType"`
	AmountLimit       int        `db:"amount_limit"       json:"amountLimit"`
	AmountObserved    int        `db:"amount_observed"    json:"amountObserved"`
	Status            string     `db:"status"             json:"status"`
	ApprovalUUID      *string    `db:"approval_uuid"      json:"approvalUuid"`
	ResolvedAt        *time.Time `db:"resolved_at"        json:"resolvedAt"`
	CreatedAt         time.Time  `db:"created_at"         json:"createdAt"`
	UpdatedAt         time.Time  `db:"updated_at"         json:"updatedAt"`
}

// CostEvent records a single AI inference cost.
type CostEvent struct {
	UUID             string    `db:"uuid"               json:"uuid"`
	CompanyUUID      string    `db:"company_uuid"       json:"companyUuid"`
	AgentUUID        string    `db:"agent_uuid"         json:"agentUuid"`
	IssueUUID        *string   `db:"issue_uuid"         json:"issueUuid"`
	ProjectUUID      *string   `db:"project_uuid"       json:"projectUuid"`
	GoalUUID         *string   `db:"goal_uuid"          json:"goalUuid"`
	HeartbeatRunUUID *string   `db:"heartbeat_run_uuid" json:"heartbeatRunUuid"`
	BillingCode      *string   `db:"billing_code"       json:"billingCode"`
	Provider         string    `db:"provider"           json:"provider"`
	Biller           string    `db:"biller"             json:"biller"`
	BillingType      string    `db:"billing_type"       json:"billingType"`
	Model            string    `db:"model"              json:"model"`
	InputTokens      int       `db:"input_tokens"       json:"inputTokens"`
	CachedInputTokens int      `db:"cached_input_tokens" json:"cachedInputTokens"`
	OutputTokens     int       `db:"output_tokens"      json:"outputTokens"`
	CostCents        int       `db:"cost_cents"         json:"costCents"`
	OccurredAt       time.Time `db:"occurred_at"        json:"occurredAt"`
	CreatedAt        time.Time `db:"created_at"         json:"createdAt"`
}

// FinanceEvent records a financial transaction (billing event).
type FinanceEvent struct {
	UUID                  string          `db:"uuid"                    json:"uuid"`
	CompanyUUID           string          `db:"company_uuid"            json:"companyUuid"`
	AgentUUID             *string         `db:"agent_uuid"              json:"agentUuid"`
	IssueUUID             *string         `db:"issue_uuid"              json:"issueUuid"`
	ProjectUUID           *string         `db:"project_uuid"            json:"projectUuid"`
	GoalUUID              *string         `db:"goal_uuid"               json:"goalUuid"`
	HeartbeatRunUUID      *string         `db:"heartbeat_run_uuid"      json:"heartbeatRunUuid"`
	CostEventUUID         *string         `db:"cost_event_uuid"         json:"costEventUuid"`
	BillingCode           *string         `db:"billing_code"            json:"billingCode"`
	Description           *string         `db:"description"             json:"description"`
	EventKind             string          `db:"event_kind"              json:"eventKind"`
	Direction             string          `db:"direction"               json:"direction"`
	Biller                string          `db:"biller"                  json:"biller"`
	Provider              *string         `db:"provider"                json:"provider"`
	ExecutionAdapterType  *string         `db:"execution_adapter_type"  json:"executionAdapterType"`
	PricingTier           *string         `db:"pricing_tier"            json:"pricingTier"`
	Region                *string         `db:"region"                  json:"region"`
	Model                 *string         `db:"model"                   json:"model"`
	Quantity              *int            `db:"quantity"                json:"quantity"`
	Unit                  *string         `db:"unit"                    json:"unit"`
	AmountCents           int             `db:"amount_cents"            json:"amountCents"`
	Currency              string          `db:"currency"                json:"currency"`
	Estimated             bool            `db:"estimated"               json:"estimated"`
	ExternalInvoiceID     *string         `db:"external_invoice_id"     json:"externalInvoiceId"`
	MetadataJSON          json.RawMessage `db:"metadata_json"           json:"metadataJson"`
	OccurredAt            time.Time       `db:"occurred_at"             json:"occurredAt"`
	CreatedAt             time.Time       `db:"created_at"              json:"createdAt"`
}
