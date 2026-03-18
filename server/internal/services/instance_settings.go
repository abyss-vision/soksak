package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/middleware"
)

// InstanceSettings represents the singleton settings row.
type InstanceSettings struct {
	UUID         string          `db:"uuid"          json:"uuid"`
	SingletonKey string          `db:"singleton_key" json:"singletonKey"`
	Experimental json.RawMessage `db:"experimental"  json:"experimental"`
	CreatedAt    time.Time       `db:"created_at"    json:"createdAt"`
	UpdatedAt    time.Time       `db:"updated_at"    json:"updatedAt"`
}

// InstanceSettingsService provides access to the singleton instance_settings row.
type InstanceSettingsService struct {
	db *sqlx.DB
}

// NewInstanceSettingsService creates a new InstanceSettingsService.
func NewInstanceSettingsService(db *sqlx.DB) *InstanceSettingsService {
	return &InstanceSettingsService{db: db}
}

// Get returns the singleton instance settings, creating it if it does not exist.
func (s *InstanceSettingsService) Get(ctx context.Context) (*InstanceSettings, error) {
	var row InstanceSettings
	err := s.db.GetContext(ctx, &row, `
		SELECT uuid, singleton_key, experimental, created_at, updated_at
		FROM instance_settings
		WHERE singleton_key = 'default'
	`)
	if errors.Is(err, sql.ErrNoRows) {
		return s.ensureDefault(ctx)
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Update merges the provided patch into the experimental JSONB field.
func (s *InstanceSettingsService) Update(ctx context.Context, patch map[string]interface{}) (*InstanceSettings, error) {
	existing, err := s.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Merge patch into existing experimental map.
	var current map[string]interface{}
	if len(existing.Experimental) > 0 {
		if err := json.Unmarshal(existing.Experimental, &current); err != nil {
			current = make(map[string]interface{})
		}
	} else {
		current = make(map[string]interface{})
	}
	for k, v := range patch {
		current[k] = v
	}

	merged, err := json.Marshal(current)
	if err != nil {
		return nil, middleware.ErrInternal("Failed to marshal settings")
	}

	var updated InstanceSettings
	err = s.db.GetContext(ctx, &updated, `
		UPDATE instance_settings
		SET experimental = $1,
		    updated_at   = now()
		WHERE uuid = $2
		RETURNING uuid, singleton_key, experimental, created_at, updated_at
	`, merged, existing.UUID)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// GetCommunicationLanguage reads communication_language from the experimental
// JSONB field. Returns empty string when not set.
func (s *InstanceSettingsService) GetCommunicationLanguage(ctx context.Context) (string, error) {
	settings, err := s.Get(ctx)
	if err != nil {
		return "", err
	}
	if len(settings.Experimental) == 0 {
		return "", nil
	}
	var exp map[string]interface{}
	if err := json.Unmarshal(settings.Experimental, &exp); err != nil {
		return "", nil
	}
	lang, _ := exp["communicationLanguage"].(string)
	return lang, nil
}

// ensureDefault inserts the default singleton row and returns it.
func (s *InstanceSettingsService) ensureDefault(ctx context.Context) (*InstanceSettings, error) {
	id := uuid.New().String()
	var row InstanceSettings
	err := s.db.GetContext(ctx, &row, `
		INSERT INTO instance_settings (uuid, singleton_key, experimental)
		VALUES ($1, 'default', '{}'::jsonb)
		ON CONFLICT (singleton_key) DO UPDATE
		    SET updated_at = instance_settings.updated_at
		RETURNING uuid, singleton_key, experimental, created_at, updated_at
	`, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}
