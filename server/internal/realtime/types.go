package realtime

import (
	"encoding/json"
	"time"
)

// WebSocketMessage is the envelope for all messages sent over the wire.
type WebSocketMessage struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	CompanyID string          `json:"companyId,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// Downstream event types (server → client).
const (
	// Run lifecycle events (heartbeat.*).
	EventHeartbeatRunQueued  = "heartbeat.run.queued"
	EventHeartbeatRunStatus  = "heartbeat.run.status"
	EventHeartbeatRunEvent   = "heartbeat.run.event"
	EventHeartbeatRunLog     = "heartbeat.run.log"
	EventRunCompleted        = "run.completed"

	// Agent events.
	EventAgentStatus        = "agent.status"
	EventAgentStatusChanged = "agent.status_changed"

	// Issue/Kanban events.
	EventIssueUpdated      = "issue.updated"
	EventKanbanIssueMoved  = "kanban.issue.moved"

	// Approval events.
	EventApprovalCreated = "approval.created"
	EventApprovalDecided = "approval.decided"
	EventApprovalUpdated = "approval.updated"

	// Misc downstream events.
	EventActivityLogged       = "activity.logged"
	EventPluginUIUpdated      = "plugin.ui.updated"
	EventPluginWorkerCrashed  = "plugin.worker.crashed"
	EventPluginWorkerRestarted = "plugin.worker.restarted"
)

// Upstream command types (client → server).
const (
	CmdAgentStdinWrite   = "agent.stdin.write"
	CmdAgentStdinSignal  = "agent.stdin.signal"
	CmdIssueUpdate       = "issue.update"
	CmdSubscribe         = "subscribe"
	CmdUnsubscribe       = "unsubscribe"
)

// StdinWritePayload carries data to write to a running process's stdin.
type StdinWritePayload struct {
	RunID string `json:"runId"`
	Data  string `json:"data"`
}

// StdinSignalPayload carries a signal to send to a running process.
type StdinSignalPayload struct {
	RunID  string `json:"runId"`
	Signal string `json:"signal"` // e.g. "SIGTERM", "SIGKILL", "SIGINT"
}

// SubscribePayload selects a channel to subscribe/unsubscribe from.
type SubscribePayload struct {
	Channel string `json:"channel"`
}

// RunLogPayload is emitted per line of stdout/stderr output.
type RunLogPayload struct {
	RunID  string `json:"runId"`
	Stream string `json:"stream"` // "stdout" or "stderr"
	Line   string `json:"line"`
}

// RunCompletedPayload is emitted when a process exits.
type RunCompletedPayload struct {
	RunID    string `json:"runId"`
	ExitCode int    `json:"exitCode"`
}
