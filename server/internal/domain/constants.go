package domain

// CompanyStatus represents the lifecycle state of a company.
type CompanyStatus string

const (
	CompanyStatusActive  CompanyStatus = "active"
	CompanyStatusPaused  CompanyStatus = "paused"
	CompanyStatusDeleted CompanyStatus = "deleted"
)

// AgentStatus represents the lifecycle state of an agent.
type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusRunning AgentStatus = "running"
	AgentStatusPaused  AgentStatus = "paused"
	AgentStatusDeleted AgentStatus = "deleted"
)

// IssueStatus represents the workflow state of an issue.
type IssueStatus string

const (
	IssueStatusBacklog    IssueStatus = "backlog"
	IssueStatusTodo       IssueStatus = "todo"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusInReview   IssueStatus = "in_review"
	IssueStatusBlocked    IssueStatus = "blocked"
	IssueStatusDone       IssueStatus = "done"
	IssueStatusCancelled  IssueStatus = "cancelled"
)

// Priority represents the priority level of work items.
type Priority string

const (
	PriorityUrgent Priority = "urgent"
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
	PriorityNone   Priority = "none"
)

// AdapterType represents the execution adapter type for agents.
type AdapterType string

const (
	AdapterTypeProcess AdapterType = "process"
	AdapterTypeDocker  AdapterType = "docker"
	AdapterTypeRemote  AdapterType = "remote"
)

// ProjectStatus represents the lifecycle state of a project.
type ProjectStatus string

const (
	ProjectStatusBacklog    ProjectStatus = "backlog"
	ProjectStatusActive     ProjectStatus = "active"
	ProjectStatusPaused     ProjectStatus = "paused"
	ProjectStatusCompleted  ProjectStatus = "completed"
	ProjectStatusCancelled  ProjectStatus = "cancelled"
)

// GoalStatus represents the lifecycle state of a goal.
type GoalStatus string

const (
	GoalStatusPlanned   GoalStatus = "planned"
	GoalStatusActive    GoalStatus = "active"
	GoalStatusCompleted GoalStatus = "completed"
	GoalStatusCancelled GoalStatus = "cancelled"
)

// GoalLevel represents the hierarchy level of a goal.
type GoalLevel string

const (
	GoalLevelTask      GoalLevel = "task"
	GoalLevelSprint    GoalLevel = "sprint"
	GoalLevelQuarter   GoalLevel = "quarter"
	GoalLevelYear      GoalLevel = "year"
)

// ApprovalStatus represents the decision state of an approval request.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

// HeartbeatRunStatus represents the state of an agent run.
type HeartbeatRunStatus string

const (
	HeartbeatRunStatusQueued    HeartbeatRunStatus = "queued"
	HeartbeatRunStatusRunning   HeartbeatRunStatus = "running"
	HeartbeatRunStatusCompleted HeartbeatRunStatus = "completed"
	HeartbeatRunStatusFailed    HeartbeatRunStatus = "failed"
	HeartbeatRunStatusCancelled HeartbeatRunStatus = "cancelled"
)

// WakeupRequestStatus represents the state of an agent wakeup request.
type WakeupRequestStatus string

const (
	WakeupRequestStatusQueued    WakeupRequestStatus = "queued"
	WakeupRequestStatusClaimed   WakeupRequestStatus = "claimed"
	WakeupRequestStatusCompleted WakeupRequestStatus = "completed"
	WakeupRequestStatusFailed    WakeupRequestStatus = "failed"
)

// PluginStatus represents the installation state of a plugin.
type PluginStatus string

const (
	PluginStatusInstalled PluginStatus = "installed"
	PluginStatusActive    PluginStatus = "active"
	PluginStatusDisabled  PluginStatus = "disabled"
	PluginStatusError     PluginStatus = "error"
)

// PluginJobStatus represents the scheduling state of a plugin job.
type PluginJobStatus string

const (
	PluginJobStatusActive PluginJobStatus = "active"
	PluginJobStatusPaused PluginJobStatus = "paused"
	PluginJobStatusError  PluginJobStatus = "error"
)

// MembershipStatus represents the state of a company membership.
type MembershipStatus string

const (
	MembershipStatusActive   MembershipStatus = "active"
	MembershipStatusInactive MembershipStatus = "inactive"
	MembershipStatusBanned   MembershipStatus = "banned"
)

// BudgetIncidentStatus represents the state of a budget incident.
type BudgetIncidentStatus string

const (
	BudgetIncidentStatusOpen      BudgetIncidentStatus = "open"
	BudgetIncidentStatusResolved  BudgetIncidentStatus = "resolved"
	BudgetIncidentStatusDismissed BudgetIncidentStatus = "dismissed"
)

// ExecutionWorkspaceStatus represents the state of an execution workspace.
type ExecutionWorkspaceStatus string

const (
	ExecutionWorkspaceStatusActive  ExecutionWorkspaceStatus = "active"
	ExecutionWorkspaceStatusClosed  ExecutionWorkspaceStatus = "closed"
	ExecutionWorkspaceStatusExpired ExecutionWorkspaceStatus = "expired"
)

// ActorType represents the type of actor performing an action.
type ActorType string

const (
	ActorTypeSystem ActorType = "system"
	ActorTypeAgent  ActorType = "agent"
	ActorTypeUser   ActorType = "user"
)
