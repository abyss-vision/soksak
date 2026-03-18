package domain

import (
	"encoding/json"
	"time"
)

// Company represents a tenant organization.
type Company struct {
	UUID                              string          `db:"uuid"                                  json:"uuid"`
	Name                              string          `db:"name"                                  json:"name"`
	Description                       *string         `db:"description"                           json:"description"`
	Status                            string          `db:"status"                                json:"status"`
	PauseReason                       *string         `db:"pause_reason"                          json:"pauseReason"`
	PausedAt                          *time.Time      `db:"paused_at"                             json:"pausedAt"`
	IssuePrefix                       string          `db:"issue_prefix"                          json:"issuePrefix"`
	IssueCounter                      int             `db:"issue_counter"                         json:"issueCounter"`
	BudgetMonthlyCents                int             `db:"budget_monthly_cents"                  json:"budgetMonthlyCents"`
	SpentMonthlyCents                 int             `db:"spent_monthly_cents"                   json:"spentMonthlyCents"`
	RequireBoardApprovalForNewAgents  bool            `db:"require_board_approval_for_new_agents" json:"requireBoardApprovalForNewAgents"`
	BrandColor                        *string         `db:"brand_color"                           json:"brandColor"`
	CommunicationLanguage             *string         `db:"communication_language"                json:"communicationLanguage"`
	CreatedAt                         time.Time       `db:"created_at"                            json:"createdAt"`
	UpdatedAt                         time.Time       `db:"updated_at"                            json:"updatedAt"`
}

// Agent represents an AI agent within a company.
type Agent struct {
	UUID                string          `db:"uuid"                  json:"uuid"`
	CompanyUUID         string          `db:"company_uuid"          json:"companyUuid"`
	Name                string          `db:"name"                  json:"name"`
	Role                string          `db:"role"                  json:"role"`
	Title               *string         `db:"title"                 json:"title"`
	Icon                *string         `db:"icon"                  json:"icon"`
	Status              string          `db:"status"                json:"status"`
	ReportsTo           *string         `db:"reports_to"            json:"reportsTo"`
	Capabilities        *string         `db:"capabilities"          json:"capabilities"`
	AdapterType         string          `db:"adapter_type"          json:"adapterType"`
	AdapterConfig       json.RawMessage `db:"adapter_config"        json:"adapterConfig"`
	RuntimeConfig       json.RawMessage `db:"runtime_config"        json:"runtimeConfig"`
	BudgetMonthlyCents  int             `db:"budget_monthly_cents"  json:"budgetMonthlyCents"`
	SpentMonthlyCents   int             `db:"spent_monthly_cents"   json:"spentMonthlyCents"`
	PauseReason         *string         `db:"pause_reason"          json:"pauseReason"`
	PausedAt            *time.Time      `db:"paused_at"             json:"pausedAt"`
	Permissions         json.RawMessage `db:"permissions"           json:"permissions"`
	LastHeartbeatAt     *time.Time      `db:"last_heartbeat_at"     json:"lastHeartbeatAt"`
	Metadata            json.RawMessage `db:"metadata"              json:"metadata"`
	CreatedAt           time.Time       `db:"created_at"            json:"createdAt"`
	UpdatedAt           time.Time       `db:"updated_at"            json:"updatedAt"`
}

// Issue represents a unit of work tracked within a company.
type Issue struct {
	UUID                          string          `db:"uuid"                            json:"uuid"`
	CompanyUUID                   string          `db:"company_uuid"                    json:"companyUuid"`
	ProjectUUID                   *string         `db:"project_uuid"                    json:"projectUuid"`
	ProjectWorkspaceUUID          *string         `db:"project_workspace_uuid"          json:"projectWorkspaceUuid"`
	GoalUUID                      *string         `db:"goal_uuid"                       json:"goalUuid"`
	ParentUUID                    *string         `db:"parent_uuid"                     json:"parentUuid"`
	Title                         string          `db:"title"                           json:"title"`
	Description                   *string         `db:"description"                     json:"description"`
	Status                        string          `db:"status"                          json:"status"`
	Priority                      string          `db:"priority"                        json:"priority"`
	AssigneeAgentUUID             *string         `db:"assignee_agent_uuid"             json:"assigneeAgentUuid"`
	AssigneeUserID                *string         `db:"assignee_user_id"                json:"assigneeUserId"`
	CheckoutRunUUID               *string         `db:"checkout_run_uuid"               json:"checkoutRunUuid"`
	ExecutionRunUUID              *string         `db:"execution_run_uuid"              json:"executionRunUuid"`
	ExecutionAgentNameKey         *string         `db:"execution_agent_name_key"        json:"executionAgentNameKey"`
	ExecutionLockedAt             *time.Time      `db:"execution_locked_at"             json:"executionLockedAt"`
	CreatedByAgentUUID            *string         `db:"created_by_agent_uuid"           json:"createdByAgentUuid"`
	CreatedByUserID               *string         `db:"created_by_user_id"              json:"createdByUserId"`
	IssueNumber                   *int            `db:"issue_number"                    json:"issueNumber"`
	Identifier                    *string         `db:"identifier"                      json:"identifier"`
	RequestDepth                  int             `db:"request_depth"                   json:"requestDepth"`
	BillingCode                   *string         `db:"billing_code"                    json:"billingCode"`
	AssigneeAdapterOverrides      json.RawMessage `db:"assignee_adapter_overrides"      json:"assigneeAdapterOverrides"`
	ExecutionWorkspaceUUID        *string         `db:"execution_workspace_uuid"        json:"executionWorkspaceUuid"`
	ExecutionWorkspacePreference  *string         `db:"execution_workspace_preference"  json:"executionWorkspacePreference"`
	ExecutionWorkspaceSettings    json.RawMessage `db:"execution_workspace_settings"    json:"executionWorkspaceSettings"`
	HiddenAt                      *time.Time      `db:"hidden_at"                       json:"hiddenAt"`
	StartedAt                     *time.Time      `db:"started_at"                      json:"startedAt"`
	CompletedAt                   *time.Time      `db:"completed_at"                    json:"completedAt"`
	CancelledAt                   *time.Time      `db:"cancelled_at"                    json:"cancelledAt"`
	CreatedAt                     time.Time       `db:"created_at"                      json:"createdAt"`
	UpdatedAt                     time.Time       `db:"updated_at"                      json:"updatedAt"`
}

// Project represents a collection of work within a company.
type Project struct {
	UUID                     string          `db:"uuid"                       json:"uuid"`
	CompanyUUID              string          `db:"company_uuid"               json:"companyUuid"`
	GoalUUID                 *string         `db:"goal_uuid"                  json:"goalUuid"`
	Name                     string          `db:"name"                       json:"name"`
	Description              *string         `db:"description"                json:"description"`
	Status                   string          `db:"status"                     json:"status"`
	LeadAgentUUID            *string         `db:"lead_agent_uuid"            json:"leadAgentUuid"`
	TargetDate               *string         `db:"target_date"                json:"targetDate"`
	Color                    *string         `db:"color"                      json:"color"`
	PauseReason              *string         `db:"pause_reason"               json:"pauseReason"`
	PausedAt                 *time.Time      `db:"paused_at"                  json:"pausedAt"`
	ExecutionWorkspacePolicy json.RawMessage `db:"execution_workspace_policy" json:"executionWorkspacePolicy"`
	ArchivedAt               *time.Time      `db:"archived_at"                json:"archivedAt"`
	CreatedAt                time.Time       `db:"created_at"                 json:"createdAt"`
	UpdatedAt                time.Time       `db:"updated_at"                 json:"updatedAt"`
}

// IssueComment represents a comment on an issue.
type IssueComment struct {
	UUID            string    `db:"uuid"              json:"uuid"`
	CompanyUUID     string    `db:"company_uuid"      json:"companyUuid"`
	IssueUUID       string    `db:"issue_uuid"        json:"issueUuid"`
	AuthorAgentUUID *string   `db:"author_agent_uuid" json:"authorAgentUuid"`
	AuthorUserID    *string   `db:"author_user_id"    json:"authorUserId"`
	Body            string    `db:"body"              json:"body"`
	CreatedAt       time.Time `db:"created_at"        json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at"        json:"updatedAt"`
}

// ProjectWorkspace represents a code workspace associated with a project.
type ProjectWorkspace struct {
	UUID                string          `db:"uuid"                   json:"uuid"`
	CompanyUUID         string          `db:"company_uuid"           json:"companyUuid"`
	ProjectUUID         string          `db:"project_uuid"           json:"projectUuid"`
	Name                string          `db:"name"                   json:"name"`
	SourceType          string          `db:"source_type"            json:"sourceType"`
	Cwd                 *string         `db:"cwd"                    json:"cwd"`
	RepoURL             *string         `db:"repo_url"               json:"repoUrl"`
	RepoRef             *string         `db:"repo_ref"               json:"repoRef"`
	DefaultRef          *string         `db:"default_ref"            json:"defaultRef"`
	Visibility          string          `db:"visibility"             json:"visibility"`
	SetupCommand        *string         `db:"setup_command"          json:"setupCommand"`
	CleanupCommand      *string         `db:"cleanup_command"        json:"cleanupCommand"`
	RemoteProvider      *string         `db:"remote_provider"        json:"remoteProvider"`
	RemoteWorkspaceRef  *string         `db:"remote_workspace_ref"   json:"remoteWorkspaceRef"`
	SharedWorkspaceKey  *string         `db:"shared_workspace_key"   json:"sharedWorkspaceKey"`
	Metadata            json.RawMessage `db:"metadata"               json:"metadata"`
	IsPrimary           bool            `db:"is_primary"             json:"isPrimary"`
	CreatedAt           time.Time       `db:"created_at"             json:"createdAt"`
	UpdatedAt           time.Time       `db:"updated_at"             json:"updatedAt"`
}

// Label represents a tag that can be applied to issues.
type Label struct {
	UUID        string    `db:"uuid"         json:"uuid"`
	CompanyUUID string    `db:"company_uuid" json:"companyUuid"`
	Name        string    `db:"name"         json:"name"`
	Color       string    `db:"color"        json:"color"`
	CreatedAt   time.Time `db:"created_at"   json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updatedAt"`
}

// IssueLabel is a join table between issues and labels.
type IssueLabel struct {
	IssueUUID   string    `db:"issue_uuid"   json:"issueUuid"`
	LabelUUID   string    `db:"label_uuid"   json:"labelUuid"`
	CompanyUUID string    `db:"company_uuid" json:"companyUuid"`
	CreatedAt   time.Time `db:"created_at"   json:"createdAt"`
}

// IssueReadState tracks the last read timestamp per user per issue.
type IssueReadState struct {
	UUID        string    `db:"uuid"         json:"uuid"`
	CompanyUUID string    `db:"company_uuid" json:"companyUuid"`
	IssueUUID   string    `db:"issue_uuid"   json:"issueUuid"`
	UserID      string    `db:"user_id"      json:"userId"`
	LastReadAt  time.Time `db:"last_read_at" json:"lastReadAt"`
	CreatedAt   time.Time `db:"created_at"   json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updatedAt"`
}
