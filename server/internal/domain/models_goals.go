package domain

import (
	"encoding/json"
	"time"
)

// Goal represents a strategic objective within a company.
type Goal struct {
	UUID          string    `db:"uuid"            json:"uuid"`
	CompanyUUID   string    `db:"company_uuid"    json:"companyUuid"`
	Title         string    `db:"title"           json:"title"`
	Description   *string   `db:"description"     json:"description"`
	Level         string    `db:"level"           json:"level"`
	Status        string    `db:"status"          json:"status"`
	ParentUUID    *string   `db:"parent_uuid"     json:"parentUuid"`
	OwnerAgentUUID *string  `db:"owner_agent_uuid" json:"ownerAgentUuid"`
	CreatedAt     time.Time `db:"created_at"      json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at"      json:"updatedAt"`
}

// ProjectGoal is a join table linking projects to goals.
type ProjectGoal struct {
	ProjectUUID string    `db:"project_uuid" json:"projectUuid"`
	GoalUUID    string    `db:"goal_uuid"    json:"goalUuid"`
	CompanyUUID string    `db:"company_uuid" json:"companyUuid"`
	CreatedAt   time.Time `db:"created_at"   json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updatedAt"`
}

// Approval represents a request for human approval of an agent action.
type Approval struct {
	UUID                  string          `db:"uuid"                     json:"uuid"`
	CompanyUUID           string          `db:"company_uuid"             json:"companyUuid"`
	Type                  string          `db:"type"                     json:"type"`
	RequestedByAgentUUID  *string         `db:"requested_by_agent_uuid"  json:"requestedByAgentUuid"`
	RequestedByUserID     *string         `db:"requested_by_user_id"     json:"requestedByUserId"`
	Status                string          `db:"status"                   json:"status"`
	Payload               json.RawMessage `db:"payload"                  json:"payload"`
	DecisionNote          *string         `db:"decision_note"            json:"decisionNote"`
	DecidedByUserID       *string         `db:"decided_by_user_id"       json:"decidedByUserId"`
	DecidedAt             *time.Time      `db:"decided_at"               json:"decidedAt"`
	CreatedAt             time.Time       `db:"created_at"               json:"createdAt"`
	UpdatedAt             time.Time       `db:"updated_at"               json:"updatedAt"`
}

// ApprovalComment represents a comment on an approval request.
type ApprovalComment struct {
	UUID             string    `db:"uuid"              json:"uuid"`
	CompanyUUID      string    `db:"company_uuid"      json:"companyUuid"`
	ApprovalUUID     string    `db:"approval_uuid"     json:"approvalUuid"`
	AuthorAgentUUID  *string   `db:"author_agent_uuid" json:"authorAgentUuid"`
	AuthorUserID     *string   `db:"author_user_id"    json:"authorUserId"`
	Body             string    `db:"body"              json:"body"`
	CreatedAt        time.Time `db:"created_at"        json:"createdAt"`
	UpdatedAt        time.Time `db:"updated_at"        json:"updatedAt"`
}

// IssueApproval links an issue to an approval request.
type IssueApproval struct {
	CompanyUUID        string    `db:"company_uuid"          json:"companyUuid"`
	IssueUUID          string    `db:"issue_uuid"            json:"issueUuid"`
	ApprovalUUID       string    `db:"approval_uuid"         json:"approvalUuid"`
	LinkedByAgentUUID  *string   `db:"linked_by_agent_uuid"  json:"linkedByAgentUuid"`
	LinkedByUserID     *string   `db:"linked_by_user_id"     json:"linkedByUserId"`
	CreatedAt          time.Time `db:"created_at"            json:"createdAt"`
}
