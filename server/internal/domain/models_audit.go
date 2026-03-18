package domain

import (
	"encoding/json"
	"time"
)

// ActivityLog records an action performed in the system.
type ActivityLog struct {
	UUID        string          `db:"uuid"        json:"uuid"`
	CompanyUUID string          `db:"company_uuid" json:"companyUuid"`
	ActorType   string          `db:"actor_type"  json:"actorType"`
	ActorID     string          `db:"actor_id"    json:"actorId"`
	Action      string          `db:"action"      json:"action"`
	EntityType  string          `db:"entity_type" json:"entityType"`
	EntityID    string          `db:"entity_id"   json:"entityId"`
	AgentUUID   *string         `db:"agent_uuid"  json:"agentUuid"`
	RunUUID     *string         `db:"run_uuid"    json:"runUuid"`
	Details     json.RawMessage `db:"details"     json:"details"`
	CreatedAt   time.Time       `db:"created_at"  json:"createdAt"`
}

// AgentConfigRevision records a change to an agent's configuration.
type AgentConfigRevision struct {
	UUID                       string          `db:"uuid"                          json:"uuid"`
	CompanyUUID                string          `db:"company_uuid"                  json:"companyUuid"`
	AgentUUID                  string          `db:"agent_uuid"                    json:"agentUuid"`
	CreatedByAgentUUID         *string         `db:"created_by_agent_uuid"         json:"createdByAgentUuid"`
	CreatedByUserID            *string         `db:"created_by_user_id"            json:"createdByUserId"`
	Source                     string          `db:"source"                        json:"source"`
	RolledBackFromRevisionUUID *string         `db:"rolled_back_from_revision_uuid" json:"rolledBackFromRevisionUuid"`
	ChangedKeys                json.RawMessage `db:"changed_keys"                  json:"changedKeys"`
	BeforeConfig               json.RawMessage `db:"before_config"                 json:"beforeConfig"`
	AfterConfig                json.RawMessage `db:"after_config"                  json:"afterConfig"`
	CreatedAt                  time.Time       `db:"created_at"                    json:"createdAt"`
}

// PrincipalPermissionGrant grants a specific permission to a principal.
type PrincipalPermissionGrant struct {
	UUID            string          `db:"uuid"              json:"uuid"`
	CompanyUUID     string          `db:"company_uuid"      json:"companyUuid"`
	PrincipalType   string          `db:"principal_type"    json:"principalType"`
	PrincipalID     string          `db:"principal_id"      json:"principalId"`
	PermissionKey   string          `db:"permission_key"    json:"permissionKey"`
	Scope           json.RawMessage `db:"scope"             json:"scope"`
	GrantedByUserID *string         `db:"granted_by_user_id" json:"grantedByUserId"`
	CreatedAt       time.Time       `db:"created_at"        json:"createdAt"`
	UpdatedAt       time.Time       `db:"updated_at"        json:"updatedAt"`
}

// InstanceSettings stores singleton instance-wide settings.
type InstanceSettings struct {
	UUID         string          `db:"uuid"          json:"uuid"`
	SingletonKey string          `db:"singleton_key" json:"singletonKey"`
	Experimental json.RawMessage `db:"experimental"  json:"experimental"`
	CreatedAt    time.Time       `db:"created_at"    json:"createdAt"`
	UpdatedAt    time.Time       `db:"updated_at"    json:"updatedAt"`
}

// CompanySecret stores a named secret for a company.
type CompanySecret struct {
	UUID                string     `db:"uuid"                  json:"uuid"`
	CompanyUUID         string     `db:"company_uuid"          json:"companyUuid"`
	Name                string     `db:"name"                  json:"name"`
	Provider            string     `db:"provider"              json:"provider"`
	ExternalRef         *string    `db:"external_ref"          json:"externalRef"`
	LatestVersion       int        `db:"latest_version"        json:"latestVersion"`
	Description         *string    `db:"description"           json:"description"`
	CreatedByAgentUUID  *string    `db:"created_by_agent_uuid" json:"createdByAgentUuid"`
	CreatedByUserID     *string    `db:"created_by_user_id"    json:"createdByUserId"`
	CreatedAt           time.Time  `db:"created_at"            json:"createdAt"`
	UpdatedAt           time.Time  `db:"updated_at"            json:"updatedAt"`
}

// CompanySecretVersion stores a versioned secret material.
type CompanySecretVersion struct {
	UUID                string          `db:"uuid"                  json:"uuid"`
	SecretUUID          string          `db:"secret_uuid"           json:"secretUuid"`
	Version             int             `db:"version"               json:"version"`
	Material            json.RawMessage `db:"material"              json:"material"`
	ValueSHA256         string          `db:"value_sha256"          json:"valueSha256"`
	CreatedByAgentUUID  *string         `db:"created_by_agent_uuid" json:"createdByAgentUuid"`
	CreatedByUserID     *string         `db:"created_by_user_id"    json:"createdByUserId"`
	CreatedAt           time.Time       `db:"created_at"            json:"createdAt"`
	RevokedAt           *time.Time      `db:"revoked_at"            json:"revokedAt"`
}

// Asset represents an uploaded file stored by a cloud provider.
type Asset struct {
	UUID                string    `db:"uuid"                  json:"uuid"`
	CompanyUUID         string    `db:"company_uuid"          json:"companyUuid"`
	Provider            string    `db:"provider"              json:"provider"`
	ObjectKey           string    `db:"object_key"            json:"objectKey"`
	ContentType         string    `db:"content_type"          json:"contentType"`
	ByteSize            int       `db:"byte_size"             json:"byteSize"`
	SHA256              string    `db:"sha256"                json:"sha256"`
	OriginalFilename    *string   `db:"original_filename"     json:"originalFilename"`
	CreatedByAgentUUID  *string   `db:"created_by_agent_uuid" json:"createdByAgentUuid"`
	CreatedByUserID     *string   `db:"created_by_user_id"    json:"createdByUserId"`
	CreatedAt           time.Time `db:"created_at"            json:"createdAt"`
	UpdatedAt           time.Time `db:"updated_at"            json:"updatedAt"`
}

// IssueAttachment links an asset to an issue (optionally to a comment).
type IssueAttachment struct {
	UUID            string    `db:"uuid"              json:"uuid"`
	CompanyUUID     string    `db:"company_uuid"      json:"companyUuid"`
	IssueUUID       string    `db:"issue_uuid"        json:"issueUuid"`
	AssetUUID       string    `db:"asset_uuid"        json:"assetUuid"`
	IssueCommentUUID *string  `db:"issue_comment_uuid" json:"issueCommentUuid"`
	CreatedAt       time.Time `db:"created_at"        json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at"        json:"updatedAt"`
}

// CompanyLogo links an asset as the logo for a company.
type CompanyLogo struct {
	UUID        string    `db:"uuid"         json:"uuid"`
	CompanyUUID string    `db:"company_uuid" json:"companyUuid"`
	AssetUUID   string    `db:"asset_uuid"   json:"assetUuid"`
	CreatedAt   time.Time `db:"created_at"   json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updatedAt"`
}

// Document represents a rich text document.
type Document struct {
	UUID                   string    `db:"uuid"                     json:"uuid"`
	CompanyUUID            string    `db:"company_uuid"             json:"companyUuid"`
	Title                  *string   `db:"title"                    json:"title"`
	Format                 string    `db:"format"                   json:"format"`
	LatestBody             string    `db:"latest_body"              json:"latestBody"`
	LatestRevisionUUID     *string   `db:"latest_revision_uuid"     json:"latestRevisionUuid"`
	LatestRevisionNumber   int       `db:"latest_revision_number"   json:"latestRevisionNumber"`
	CreatedByAgentUUID     *string   `db:"created_by_agent_uuid"    json:"createdByAgentUuid"`
	CreatedByUserID        *string   `db:"created_by_user_id"       json:"createdByUserId"`
	UpdatedByAgentUUID     *string   `db:"updated_by_agent_uuid"    json:"updatedByAgentUuid"`
	UpdatedByUserID        *string   `db:"updated_by_user_id"       json:"updatedByUserId"`
	CreatedAt              time.Time `db:"created_at"               json:"createdAt"`
	UpdatedAt              time.Time `db:"updated_at"               json:"updatedAt"`
}

// DocumentRevision records a historical version of a document.
type DocumentRevision struct {
	UUID                string    `db:"uuid"                  json:"uuid"`
	CompanyUUID         string    `db:"company_uuid"          json:"companyUuid"`
	DocumentUUID        string    `db:"document_uuid"         json:"documentUuid"`
	RevisionNumber      int       `db:"revision_number"       json:"revisionNumber"`
	Body                string    `db:"body"                  json:"body"`
	ChangeSummary       *string   `db:"change_summary"        json:"changeSummary"`
	CreatedByAgentUUID  *string   `db:"created_by_agent_uuid" json:"createdByAgentUuid"`
	CreatedByUserID     *string   `db:"created_by_user_id"    json:"createdByUserId"`
	CreatedAt           time.Time `db:"created_at"            json:"createdAt"`
}

// IssueDocument links a document to an issue under a named key.
type IssueDocument struct {
	UUID        string    `db:"uuid"         json:"uuid"`
	CompanyUUID string    `db:"company_uuid" json:"companyUuid"`
	IssueUUID   string    `db:"issue_uuid"   json:"issueUuid"`
	DocumentUUID string   `db:"document_uuid" json:"documentUuid"`
	Key         string    `db:"key"          json:"key"`
	CreatedAt   time.Time `db:"created_at"   json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updatedAt"`
}
