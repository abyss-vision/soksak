package domain

import (
	"encoding/json"
	"time"
)

// HeartbeatRun represents a single execution run of an agent.
type HeartbeatRun struct {
	UUID               string          `db:"uuid"                 json:"uuid"`
	CompanyUUID        string          `db:"company_uuid"         json:"companyUuid"`
	AgentUUID          string          `db:"agent_uuid"           json:"agentUuid"`
	InvocationSource   string          `db:"invocation_source"    json:"invocationSource"`
	TriggerDetail      *string         `db:"trigger_detail"       json:"triggerDetail"`
	Status             string          `db:"status"               json:"status"`
	StartedAt          *time.Time      `db:"started_at"           json:"startedAt"`
	FinishedAt         *time.Time      `db:"finished_at"          json:"finishedAt"`
	Error              *string         `db:"error"                json:"error"`
	WakeupRequestUUID  *string         `db:"wakeup_request_uuid"  json:"wakeupRequestUuid"`
	ExitCode           *int            `db:"exit_code"            json:"exitCode"`
	Signal             *string         `db:"signal"               json:"signal"`
	UsageJSON          json.RawMessage `db:"usage_json"           json:"usageJson"`
	ResultJSON         json.RawMessage `db:"result_json"          json:"resultJson"`
	SessionIDBefore    *string         `db:"session_id_before"    json:"sessionIdBefore"`
	SessionIDAfter     *string         `db:"session_id_after"     json:"sessionIdAfter"`
	LogStore           *string         `db:"log_store"            json:"logStore"`
	LogRef             *string         `db:"log_ref"              json:"logRef"`
	LogBytes           *int64          `db:"log_bytes"            json:"logBytes"`
	LogSHA256          *string         `db:"log_sha256"           json:"logSha256"`
	LogCompressed      bool            `db:"log_compressed"       json:"logCompressed"`
	StdoutExcerpt      *string         `db:"stdout_excerpt"       json:"stdoutExcerpt"`
	StderrExcerpt      *string         `db:"stderr_excerpt"       json:"stderrExcerpt"`
	ErrorCode          *string         `db:"error_code"           json:"errorCode"`
	ExternalRunID      *string         `db:"external_run_id"      json:"externalRunId"`
	ContextSnapshot    json.RawMessage `db:"context_snapshot"     json:"contextSnapshot"`
	CreatedAt          time.Time       `db:"created_at"           json:"createdAt"`
	UpdatedAt          time.Time       `db:"updated_at"           json:"updatedAt"`
}

// HeartbeatRunEvent is a single streaming event from an agent run.
// Uses bigserial PK — the only table with bare "id".
type HeartbeatRunEvent struct {
	ID          int64           `db:"id"           json:"id"`
	CompanyUUID string          `db:"company_uuid" json:"companyUuid"`
	RunUUID     string          `db:"run_uuid"     json:"runUuid"`
	AgentUUID   string          `db:"agent_uuid"   json:"agentUuid"`
	Seq         int             `db:"seq"          json:"seq"`
	EventType   string          `db:"event_type"   json:"eventType"`
	Stream      *string         `db:"stream"       json:"stream"`
	Level       *string         `db:"level"        json:"level"`
	Color       *string         `db:"color"        json:"color"`
	Message     *string         `db:"message"      json:"message"`
	Payload     json.RawMessage `db:"payload"      json:"payload"`
	CreatedAt   time.Time       `db:"created_at"   json:"createdAt"`
}

// WakeupRequest represents a request to wake up an agent.
type WakeupRequest struct {
	UUID                   string          `db:"uuid"                      json:"uuid"`
	CompanyUUID            string          `db:"company_uuid"              json:"companyUuid"`
	AgentUUID              string          `db:"agent_uuid"                json:"agentUuid"`
	Source                 string          `db:"source"                    json:"source"`
	TriggerDetail          *string         `db:"trigger_detail"            json:"triggerDetail"`
	Reason                 *string         `db:"reason"                    json:"reason"`
	Payload                json.RawMessage `db:"payload"                   json:"payload"`
	Status                 string          `db:"status"                    json:"status"`
	CoalescedCount         int             `db:"coalesced_count"           json:"coalescedCount"`
	RequestedByActorType   *string         `db:"requested_by_actor_type"   json:"requestedByActorType"`
	RequestedByActorID     *string         `db:"requested_by_actor_id"     json:"requestedByActorId"`
	IdempotencyKey         *string         `db:"idempotency_key"           json:"idempotencyKey"`
	RunUUID                *string         `db:"run_uuid"                  json:"runUuid"`
	RequestedAt            time.Time       `db:"requested_at"              json:"requestedAt"`
	ClaimedAt              *time.Time      `db:"claimed_at"                json:"claimedAt"`
	FinishedAt             *time.Time      `db:"finished_at"               json:"finishedAt"`
	Error                  *string         `db:"error"                     json:"error"`
	CreatedAt              time.Time       `db:"created_at"                json:"createdAt"`
	UpdatedAt              time.Time       `db:"updated_at"                json:"updatedAt"`
}

// RuntimeState stores the persistent runtime state for an agent.
// PK is agent_uuid (no separate uuid column).
type RuntimeState struct {
	AgentUUID              string          `db:"agent_uuid"               json:"agentUuid"`
	CompanyUUID            string          `db:"company_uuid"             json:"companyUuid"`
	AdapterType            string          `db:"adapter_type"             json:"adapterType"`
	SessionID              *string         `db:"session_id"               json:"sessionId"`
	StateJSON              json.RawMessage `db:"state_json"               json:"stateJson"`
	LastRunUUID            *string         `db:"last_run_uuid"            json:"lastRunUuid"`
	LastRunStatus          *string         `db:"last_run_status"          json:"lastRunStatus"`
	TotalInputTokens       int64           `db:"total_input_tokens"       json:"totalInputTokens"`
	TotalOutputTokens      int64           `db:"total_output_tokens"      json:"totalOutputTokens"`
	TotalCachedInputTokens int64           `db:"total_cached_input_tokens" json:"totalCachedInputTokens"`
	TotalCostCents         int64           `db:"total_cost_cents"         json:"totalCostCents"`
	LastError              *string         `db:"last_error"               json:"lastError"`
	CreatedAt              time.Time       `db:"created_at"               json:"createdAt"`
	UpdatedAt              time.Time       `db:"updated_at"               json:"updatedAt"`
}

// ExecutionWorkspace represents a temporary workspace for agent execution.
type ExecutionWorkspace struct {
	UUID                               string          `db:"uuid"                                       json:"uuid"`
	CompanyUUID                        string          `db:"company_uuid"                               json:"companyUuid"`
	ProjectUUID                        string          `db:"project_uuid"                               json:"projectUuid"`
	ProjectWorkspaceUUID               *string         `db:"project_workspace_uuid"                     json:"projectWorkspaceUuid"`
	SourceIssueUUID                    *string         `db:"source_issue_uuid"                          json:"sourceIssueUuid"`
	Mode                               string          `db:"mode"                                       json:"mode"`
	StrategyType                       string          `db:"strategy_type"                              json:"strategyType"`
	Name                               string          `db:"name"                                       json:"name"`
	Status                             string          `db:"status"                                     json:"status"`
	Cwd                                *string         `db:"cwd"                                        json:"cwd"`
	RepoURL                            *string         `db:"repo_url"                                   json:"repoUrl"`
	BaseRef                            *string         `db:"base_ref"                                   json:"baseRef"`
	BranchName                         *string         `db:"branch_name"                                json:"branchName"`
	ProviderType                       string          `db:"provider_type"                              json:"providerType"`
	ProviderRef                        *string         `db:"provider_ref"                               json:"providerRef"`
	DerivedFromExecutionWorkspaceUUID  *string         `db:"derived_from_execution_workspace_uuid"      json:"derivedFromExecutionWorkspaceUuid"`
	LastUsedAt                         time.Time       `db:"last_used_at"                               json:"lastUsedAt"`
	OpenedAt                           time.Time       `db:"opened_at"                                  json:"openedAt"`
	ClosedAt                           *time.Time      `db:"closed_at"                                  json:"closedAt"`
	CleanupEligibleAt                  *time.Time      `db:"cleanup_eligible_at"                        json:"cleanupEligibleAt"`
	CleanupReason                      *string         `db:"cleanup_reason"                             json:"cleanupReason"`
	Metadata                           json.RawMessage `db:"metadata"                                   json:"metadata"`
	CreatedAt                          time.Time       `db:"created_at"                                 json:"createdAt"`
	UpdatedAt                          time.Time       `db:"updated_at"                                 json:"updatedAt"`
}

// AgentTaskSession stores session state for agent task adapters.
type AgentTaskSession struct {
	UUID               string          `db:"uuid"                 json:"uuid"`
	CompanyUUID        string          `db:"company_uuid"         json:"companyUuid"`
	AgentUUID          string          `db:"agent_uuid"           json:"agentUuid"`
	AdapterType        string          `db:"adapter_type"         json:"adapterType"`
	TaskKey            string          `db:"task_key"             json:"taskKey"`
	SessionParamsJSON  json.RawMessage `db:"session_params_json"  json:"sessionParamsJson"`
	SessionDisplayID   *string         `db:"session_display_id"   json:"sessionDisplayId"`
	LastRunUUID        *string         `db:"last_run_uuid"        json:"lastRunUuid"`
	LastError          *string         `db:"last_error"           json:"lastError"`
	CreatedAt          time.Time       `db:"created_at"           json:"createdAt"`
	UpdatedAt          time.Time       `db:"updated_at"           json:"updatedAt"`
}

// WorkspaceOperation tracks individual operations within an execution workspace.
type WorkspaceOperation struct {
	UUID                   string          `db:"uuid"                      json:"uuid"`
	CompanyUUID            string          `db:"company_uuid"              json:"companyUuid"`
	ExecutionWorkspaceUUID *string         `db:"execution_workspace_uuid"  json:"executionWorkspaceUuid"`
	HeartbeatRunUUID       *string         `db:"heartbeat_run_uuid"        json:"heartbeatRunUuid"`
	Phase                  string          `db:"phase"                     json:"phase"`
	Command                *string         `db:"command"                   json:"command"`
	Cwd                    *string         `db:"cwd"                       json:"cwd"`
	Status                 string          `db:"status"                    json:"status"`
	ExitCode               *int            `db:"exit_code"                 json:"exitCode"`
	LogStore               *string         `db:"log_store"                 json:"logStore"`
	LogRef                 *string         `db:"log_ref"                   json:"logRef"`
	LogBytes               *int64          `db:"log_bytes"                 json:"logBytes"`
	LogSHA256              *string         `db:"log_sha256"                json:"logSha256"`
	LogCompressed          bool            `db:"log_compressed"            json:"logCompressed"`
	StdoutExcerpt          *string         `db:"stdout_excerpt"            json:"stdoutExcerpt"`
	StderrExcerpt          *string         `db:"stderr_excerpt"            json:"stderrExcerpt"`
	Metadata               json.RawMessage `db:"metadata"                  json:"metadata"`
	StartedAt              time.Time       `db:"started_at"                json:"startedAt"`
	FinishedAt             *time.Time      `db:"finished_at"               json:"finishedAt"`
	CreatedAt              time.Time       `db:"created_at"                json:"createdAt"`
	UpdatedAt              time.Time       `db:"updated_at"                json:"updatedAt"`
}

// WorkspaceRuntimeService represents a running service within a workspace.
type WorkspaceRuntimeService struct {
	UUID                   string          `db:"uuid"                       json:"uuid"`
	CompanyUUID            string          `db:"company_uuid"               json:"companyUuid"`
	ProjectUUID            *string         `db:"project_uuid"               json:"projectUuid"`
	ProjectWorkspaceUUID   *string         `db:"project_workspace_uuid"     json:"projectWorkspaceUuid"`
	ExecutionWorkspaceUUID *string         `db:"execution_workspace_uuid"   json:"executionWorkspaceUuid"`
	IssueUUID              *string         `db:"issue_uuid"                 json:"issueUuid"`
	ScopeType              string          `db:"scope_type"                 json:"scopeType"`
	ScopeID                *string         `db:"scope_id"                   json:"scopeId"`
	ServiceName            string          `db:"service_name"               json:"serviceName"`
	Status                 string          `db:"status"                     json:"status"`
	Lifecycle              string          `db:"lifecycle"                  json:"lifecycle"`
	ReuseKey               *string         `db:"reuse_key"                  json:"reuseKey"`
	Command                *string         `db:"command"                    json:"command"`
	Cwd                    *string         `db:"cwd"                        json:"cwd"`
	Port                   *int            `db:"port"                       json:"port"`
	URL                    *string         `db:"url"                        json:"url"`
	Provider               string          `db:"provider"                   json:"provider"`
	ProviderRef            *string         `db:"provider_ref"               json:"providerRef"`
	OwnerAgentUUID         *string         `db:"owner_agent_uuid"           json:"ownerAgentUuid"`
	StartedByRunUUID       *string         `db:"started_by_run_uuid"        json:"startedByRunUuid"`
	LastUsedAt             time.Time       `db:"last_used_at"               json:"lastUsedAt"`
	StartedAt              time.Time       `db:"started_at"                 json:"startedAt"`
	StoppedAt              *time.Time      `db:"stopped_at"                 json:"stoppedAt"`
	StopPolicy             json.RawMessage `db:"stop_policy"                json:"stopPolicy"`
	HealthStatus           string          `db:"health_status"              json:"healthStatus"`
	CreatedAt              time.Time       `db:"created_at"                 json:"createdAt"`
	UpdatedAt              time.Time       `db:"updated_at"                 json:"updatedAt"`
}

// IssueWorkProduct represents an output artifact produced by an agent for an issue.
type IssueWorkProduct struct {
	UUID                   string          `db:"uuid"                       json:"uuid"`
	CompanyUUID            string          `db:"company_uuid"               json:"companyUuid"`
	ProjectUUID            *string         `db:"project_uuid"               json:"projectUuid"`
	IssueUUID              string          `db:"issue_uuid"                 json:"issueUuid"`
	ExecutionWorkspaceUUID *string         `db:"execution_workspace_uuid"   json:"executionWorkspaceUuid"`
	RuntimeServiceUUID     *string         `db:"runtime_service_uuid"       json:"runtimeServiceUuid"`
	Type                   string          `db:"type"                       json:"type"`
	Provider               string          `db:"provider"                   json:"provider"`
	ExternalID             *string         `db:"external_id"                json:"externalId"`
	Title                  string          `db:"title"                      json:"title"`
	URL                    *string         `db:"url"                        json:"url"`
	Status                 string          `db:"status"                     json:"status"`
	ReviewState            string          `db:"review_state"               json:"reviewState"`
	IsPrimary              bool            `db:"is_primary"                 json:"isPrimary"`
	HealthStatus           string          `db:"health_status"              json:"healthStatus"`
	Summary                *string         `db:"summary"                    json:"summary"`
	Metadata               json.RawMessage `db:"metadata"                   json:"metadata"`
	CreatedByRunUUID       *string         `db:"created_by_run_uuid"        json:"createdByRunUuid"`
	CreatedAt              time.Time       `db:"created_at"                 json:"createdAt"`
	UpdatedAt              time.Time       `db:"updated_at"                 json:"updatedAt"`
}
