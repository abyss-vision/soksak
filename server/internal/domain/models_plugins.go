package domain

import (
	"encoding/json"
	"time"
)

// Plugin represents an installed plugin.
type Plugin struct {
	UUID         string          `db:"uuid"          json:"uuid"`
	PluginKey    string          `db:"plugin_key"    json:"pluginKey"`
	PackageName  string          `db:"package_name"  json:"packageName"`
	Version      string          `db:"version"       json:"version"`
	APIVersion   int             `db:"api_version"   json:"apiVersion"`
	Categories   json.RawMessage `db:"categories"    json:"categories"`
	ManifestJSON json.RawMessage `db:"manifest_json" json:"manifestJson"`
	Status       string          `db:"status"        json:"status"`
	InstallOrder *int            `db:"install_order" json:"installOrder"`
	PackagePath  *string         `db:"package_path"  json:"packagePath"`
	LastError    *string         `db:"last_error"    json:"lastError"`
	InstalledAt  time.Time       `db:"installed_at"  json:"installedAt"`
	UpdatedAt    time.Time       `db:"updated_at"    json:"updatedAt"`
}

// PluginConfig stores operator configuration for a plugin.
type PluginConfig struct {
	UUID       string          `db:"uuid"        json:"uuid"`
	PluginUUID string          `db:"plugin_uuid" json:"pluginUuid"`
	ConfigJSON json.RawMessage `db:"config_json" json:"configJson"`
	LastError  *string         `db:"last_error"  json:"lastError"`
	CreatedAt  time.Time       `db:"created_at"  json:"createdAt"`
	UpdatedAt  time.Time       `db:"updated_at"  json:"updatedAt"`
}

// PluginJob represents a scheduled job registered by a plugin.
type PluginJob struct {
	UUID       string     `db:"uuid"        json:"uuid"`
	PluginUUID string     `db:"plugin_uuid" json:"pluginUuid"`
	JobKey     string     `db:"job_key"     json:"jobKey"`
	Schedule   string     `db:"schedule"    json:"schedule"`
	Status     string     `db:"status"      json:"status"`
	LastRunAt  *time.Time `db:"last_run_at" json:"lastRunAt"`
	NextRunAt  *time.Time `db:"next_run_at" json:"nextRunAt"`
	CreatedAt  time.Time  `db:"created_at"  json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updatedAt"`
}

// PluginJobRun records a single execution of a plugin job.
type PluginJobRun struct {
	UUID       string          `db:"uuid"        json:"uuid"`
	JobUUID    string          `db:"job_uuid"    json:"jobUuid"`
	PluginUUID string          `db:"plugin_uuid" json:"pluginUuid"`
	Trigger    string          `db:"trigger"     json:"trigger"`
	Status     string          `db:"status"      json:"status"`
	DurationMs *int            `db:"duration_ms" json:"durationMs"`
	Error      *string         `db:"error"       json:"error"`
	Logs       json.RawMessage `db:"logs"        json:"logs"`
	StartedAt  *time.Time      `db:"started_at"  json:"startedAt"`
	FinishedAt *time.Time      `db:"finished_at" json:"finishedAt"`
	CreatedAt  time.Time       `db:"created_at"  json:"createdAt"`
}

// PluginWebhookDelivery records an inbound webhook delivery for a plugin.
type PluginWebhookDelivery struct {
	UUID       string          `db:"uuid"        json:"uuid"`
	PluginUUID string          `db:"plugin_uuid" json:"pluginUuid"`
	WebhookKey string          `db:"webhook_key" json:"webhookKey"`
	ExternalID *string         `db:"external_id" json:"externalId"`
	Status     string          `db:"status"      json:"status"`
	DurationMs *int            `db:"duration_ms" json:"durationMs"`
	Error      *string         `db:"error"       json:"error"`
	Payload    json.RawMessage `db:"payload"     json:"payload"`
	Headers    json.RawMessage `db:"headers"     json:"headers"`
	StartedAt  *time.Time      `db:"started_at"  json:"startedAt"`
	FinishedAt *time.Time      `db:"finished_at" json:"finishedAt"`
	CreatedAt  time.Time       `db:"created_at"  json:"createdAt"`
}

// PluginState is a scoped key-value store entry for a plugin.
type PluginState struct {
	UUID      string          `db:"uuid"       json:"uuid"`
	PluginUUID string         `db:"plugin_uuid" json:"pluginUuid"`
	ScopeKind string          `db:"scope_kind" json:"scopeKind"`
	ScopeID   *string         `db:"scope_id"   json:"scopeId"`
	Namespace string          `db:"namespace"  json:"namespace"`
	StateKey  string          `db:"state_key"  json:"stateKey"`
	ValueJSON json.RawMessage `db:"value_json" json:"valueJson"`
	UpdatedAt time.Time       `db:"updated_at" json:"updatedAt"`
}

// PluginEntity represents a mapping between a Soksak object and an external plugin entity.
type PluginEntity struct {
	UUID       string          `db:"uuid"        json:"uuid"`
	PluginUUID string          `db:"plugin_uuid" json:"pluginUuid"`
	EntityType string          `db:"entity_type" json:"entityType"`
	ScopeKind  string          `db:"scope_kind"  json:"scopeKind"`
	ScopeID    *string         `db:"scope_id"    json:"scopeId"`
	ExternalID *string         `db:"external_id" json:"externalId"`
	Title      *string         `db:"title"       json:"title"`
	Status     *string         `db:"status"      json:"status"`
	Data       json.RawMessage `db:"data"        json:"data"`
	CreatedAt  time.Time       `db:"created_at"  json:"createdAt"`
	UpdatedAt  time.Time       `db:"updated_at"  json:"updatedAt"`
}

// PluginCompanySettings stores per-company settings for a plugin.
type PluginCompanySettings struct {
	UUID         string          `db:"uuid"          json:"uuid"`
	CompanyUUID  string          `db:"company_uuid"  json:"companyUuid"`
	PluginUUID   string          `db:"plugin_uuid"   json:"pluginUuid"`
	Enabled      bool            `db:"enabled"       json:"enabled"`
	SettingsJSON json.RawMessage `db:"settings_json" json:"settingsJson"`
	LastError    *string         `db:"last_error"    json:"lastError"`
	CreatedAt    time.Time       `db:"created_at"    json:"createdAt"`
	UpdatedAt    time.Time       `db:"updated_at"    json:"updatedAt"`
}

// PluginLog is a log line emitted by a plugin worker.
type PluginLog struct {
	UUID       string          `db:"uuid"        json:"uuid"`
	PluginUUID string          `db:"plugin_uuid" json:"pluginUuid"`
	Level      string          `db:"level"       json:"level"`
	Message    string          `db:"message"     json:"message"`
	Meta       json.RawMessage `db:"meta"        json:"meta"`
	CreatedAt  time.Time       `db:"created_at"  json:"createdAt"`
}
