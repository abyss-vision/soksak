package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/domain"
)

const (
	AgentStatusIdle    = "idle"
	AgentStatusRunning = "running"
	AgentStatusPaused  = "paused"
	AgentStatusError   = "error"
	AgentStatusDeleted = "deleted"
)

// AgentService handles CRUD and lifecycle operations for agents.
type AgentService struct {
	db *sqlx.DB
}

// NewAgentService creates a new AgentService.
func NewAgentService(db *sqlx.DB) *AgentService {
	return &AgentService{db: db}
}

// List returns all agents for a company.
func (s *AgentService) List(ctx context.Context, companyUUID string) ([]domain.Agent, error) {
	var agents []domain.Agent
	err := s.db.SelectContext(ctx, &agents,
		`SELECT * FROM agents WHERE company_uuid = $1 AND status != $2 ORDER BY created_at DESC`,
		companyUUID, AgentStatusDeleted,
	)
	if err != nil {
		return nil, fmt.Errorf("agents.List: %w", err)
	}
	return agents, nil
}

// Get returns a single agent by UUID within a company.
func (s *AgentService) Get(ctx context.Context, companyUUID, agentUUID string) (*domain.Agent, error) {
	var agent domain.Agent
	err := s.db.GetContext(ctx, &agent,
		`SELECT * FROM agents WHERE uuid = $1 AND company_uuid = $2`,
		agentUUID, companyUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("agents.Get: %w", err)
	}
	return &agent, nil
}

// CreateAgentInput holds fields for creating a new agent.
type CreateAgentInput struct {
	Name               string          `json:"name"`
	Role               string          `json:"role"`
	Title              *string         `json:"title"`
	Icon               *string         `json:"icon"`
	ReportsTo          *string         `json:"reportsTo"`
	Capabilities       *string         `json:"capabilities"`
	AdapterType        string          `json:"adapterType"`
	AdapterConfig      json.RawMessage `json:"adapterConfig"`
	RuntimeConfig      json.RawMessage `json:"runtimeConfig"`
	BudgetMonthlyCents int             `json:"budgetMonthlyCents"`
	Permissions        json.RawMessage `json:"permissions"`
	Metadata           json.RawMessage `json:"metadata"`
}

// Create creates a new agent for a company.
func (s *AgentService) Create(ctx context.Context, companyUUID string, input CreateAgentInput) (*domain.Agent, error) {
	id := uuid.NewString()
	adapterConfig := input.AdapterConfig
	if len(adapterConfig) == 0 {
		adapterConfig = json.RawMessage("{}")
	}
	runtimeConfig := input.RuntimeConfig
	if len(runtimeConfig) == 0 {
		runtimeConfig = json.RawMessage("{}")
	}
	permissions := input.Permissions
	if len(permissions) == 0 {
		permissions = json.RawMessage("{}")
	}
	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO agents (
			uuid, company_uuid, name, role, title, icon, status,
			reports_to, capabilities, adapter_type, adapter_config,
			runtime_config, budget_monthly_cents, permissions, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15
		)`,
		id, companyUUID, input.Name, input.Role, input.Title, input.Icon, AgentStatusIdle,
		input.ReportsTo, input.Capabilities, input.AdapterType, adapterConfig,
		runtimeConfig, input.BudgetMonthlyCents, permissions, metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("agents.Create: %w", err)
	}
	return s.Get(ctx, companyUUID, id)
}

// UpdateAgentInput holds fields for updating an agent.
type UpdateAgentInput struct {
	Name               *string         `json:"name"`
	Role               *string         `json:"role"`
	Title              *string         `json:"title"`
	Icon               *string         `json:"icon"`
	ReportsTo          *string         `json:"reportsTo"`
	Capabilities       *string         `json:"capabilities"`
	AdapterType        *string         `json:"adapterType"`
	AdapterConfig      json.RawMessage `json:"adapterConfig"`
	RuntimeConfig      json.RawMessage `json:"runtimeConfig"`
	BudgetMonthlyCents *int            `json:"budgetMonthlyCents"`
	Permissions        json.RawMessage `json:"permissions"`
	Metadata           json.RawMessage `json:"metadata"`
}

// Update updates an agent's fields.
func (s *AgentService) Update(ctx context.Context, companyUUID, agentUUID string, input UpdateAgentInput) (*domain.Agent, error) {
	existing, err := s.Get(ctx, companyUUID, agentUUID)
	if err != nil {
		return nil, err
	}

	name := existing.Name
	if input.Name != nil {
		name = *input.Name
	}
	role := existing.Role
	if input.Role != nil {
		role = *input.Role
	}
	title := existing.Title
	if input.Title != nil {
		title = input.Title
	}
	icon := existing.Icon
	if input.Icon != nil {
		icon = input.Icon
	}
	reportsTo := existing.ReportsTo
	if input.ReportsTo != nil {
		reportsTo = input.ReportsTo
	}
	capabilities := existing.Capabilities
	if input.Capabilities != nil {
		capabilities = input.Capabilities
	}
	adapterType := existing.AdapterType
	if input.AdapterType != nil {
		adapterType = *input.AdapterType
	}
	adapterConfig := existing.AdapterConfig
	if len(input.AdapterConfig) > 0 {
		adapterConfig = input.AdapterConfig
	}
	runtimeConfig := existing.RuntimeConfig
	if len(input.RuntimeConfig) > 0 {
		runtimeConfig = input.RuntimeConfig
	}
	budget := existing.BudgetMonthlyCents
	if input.BudgetMonthlyCents != nil {
		budget = *input.BudgetMonthlyCents
	}
	permissions := existing.Permissions
	if len(input.Permissions) > 0 {
		permissions = input.Permissions
	}
	metadata := existing.Metadata
	if len(input.Metadata) > 0 {
		metadata = input.Metadata
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE agents SET
			name = $1, role = $2, title = $3, icon = $4,
			reports_to = $5, capabilities = $6,
			adapter_type = $7, adapter_config = $8, runtime_config = $9,
			budget_monthly_cents = $10, permissions = $11, metadata = $12,
			updated_at = now()
		WHERE uuid = $13 AND company_uuid = $14`,
		name, role, title, icon,
		reportsTo, capabilities,
		adapterType, adapterConfig, runtimeConfig,
		budget, permissions, metadata,
		agentUUID, companyUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("agents.Update: %w", err)
	}
	return s.Get(ctx, companyUUID, agentUUID)
}

// Delete soft-deletes an agent by setting status to deleted.
func (s *AgentService) Delete(ctx context.Context, companyUUID, agentUUID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE agents SET status = $1, updated_at = now()
		WHERE uuid = $2 AND company_uuid = $3`,
		AgentStatusDeleted, agentUUID, companyUUID,
	)
	if err != nil {
		return fmt.Errorf("agents.Delete: %w", err)
	}
	return nil
}

// Hire sets the agent status to idle and applies an optional adapter config override.
func (s *AgentService) Hire(ctx context.Context, companyUUID, agentUUID string, config json.RawMessage) error {
	query := `UPDATE agents SET status = $1, updated_at = now() WHERE uuid = $2 AND company_uuid = $3`
	args := []any{AgentStatusIdle, agentUUID, companyUUID}

	if len(config) > 0 {
		query = `UPDATE agents SET status = $1, adapter_config = $2, updated_at = now() WHERE uuid = $3 AND company_uuid = $4`
		args = []any{AgentStatusIdle, config, agentUUID, companyUUID}
	}

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("agents.Hire: %w", err)
	}
	return nil
}

// Fire soft-deletes the agent by setting status to deleted.
func (s *AgentService) Fire(ctx context.Context, companyUUID, agentUUID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE agents SET status = $1, updated_at = now()
		WHERE uuid = $2 AND company_uuid = $3`,
		AgentStatusDeleted, agentUUID, companyUUID,
	)
	if err != nil {
		return fmt.Errorf("agents.Fire: %w", err)
	}
	return nil
}

// agentRuntimeConfig is a partial view of the agent runtime_config JSON
// used for reading per-agent overrides.
type agentRuntimeConfig struct {
	CommunicationLanguage string `json:"communication_language"`
}

// ResolveCommunicationLanguage determines the effective communication language
// for an agent run using the resolution chain:
//  1. agent.runtime_config.communication_language (if non-empty)
//  2. company.communication_language (if non-nil and non-empty)
//  3. instance settings communicationLanguage (if non-empty)
//  4. "en" fallback
func (s *AgentService) ResolveCommunicationLanguage(
	ctx context.Context,
	agent *domain.Agent,
	company *domain.Company,
	instanceSettingsSvc *InstanceSettingsService,
) (string, error) {
	// 1. Agent runtime_config override.
	if len(agent.RuntimeConfig) > 0 {
		var rc agentRuntimeConfig
		if err := json.Unmarshal(agent.RuntimeConfig, &rc); err == nil && rc.CommunicationLanguage != "" {
			return rc.CommunicationLanguage, nil
		}
	}

	// 2. Company-level setting.
	if company != nil && company.CommunicationLanguage != nil && *company.CommunicationLanguage != "" {
		return *company.CommunicationLanguage, nil
	}

	// 3. Instance settings.
	if instanceSettingsSvc != nil {
		lang, err := instanceSettingsSvc.GetCommunicationLanguage(ctx)
		if err == nil && lang != "" {
			return lang, nil
		}
	}

	// 4. Default fallback.
	return "en", nil
}

// Pause sets the agent status to paused.
func (s *AgentService) Pause(ctx context.Context, companyUUID, agentUUID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE agents SET status = $1, paused_at = now(), updated_at = now()
		WHERE uuid = $2 AND company_uuid = $3`,
		AgentStatusPaused, agentUUID, companyUUID,
	)
	if err != nil {
		return fmt.Errorf("agents.Pause: %w", err)
	}
	return nil
}
