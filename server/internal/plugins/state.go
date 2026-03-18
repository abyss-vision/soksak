package plugins

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// StateStore provides a scoped key-value store backed by the plugin_state table.
// Keys are namespaced by plugin UUID, scope kind, scope ID, namespace, and key.
type StateStore struct {
	db *sqlx.DB
}

// NewStateStore creates a StateStore using the given database handle.
func NewStateStore(db *sqlx.DB) *StateStore {
	return &StateStore{db: db}
}

// StateEntry is the value returned from Get and List.
type StateEntry struct {
	UUID       string          `db:"uuid"`
	PluginUUID string          `db:"plugin_uuid"`
	ScopeKind  string          `db:"scope_kind"`
	ScopeID    *string         `db:"scope_id"`
	Namespace  string          `db:"namespace"`
	StateKey   string          `db:"state_key"`
	ValueJSON  json.RawMessage `db:"value_json"`
}

// Get retrieves the stored value for (pluginName, key) using the "global" scope
// and "default" namespace. Returns (nil, nil) when the key does not exist.
func (s *StateStore) Get(ctx context.Context, pluginUUID, key string) (json.RawMessage, error) {
	const q = `
		SELECT uuid, plugin_uuid, scope_kind, scope_id, namespace, state_key, value_json
		FROM plugin_state
		WHERE plugin_uuid = $1
		  AND scope_kind  = 'global'
		  AND scope_id    IS NULL
		  AND namespace   = 'default'
		  AND state_key   = $2`

	var entry StateEntry
	err := s.db.GetContext(ctx, &entry, q, pluginUUID, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("state get %s/%s: %w", pluginUUID, key, err)
	}
	return entry.ValueJSON, nil
}

// Set upserts a value for (pluginUUID, key) in the "global" / "default" scope.
// value must be JSON-serializable.
func (s *StateStore) Set(ctx context.Context, pluginUUID, key string, value interface{}) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal state value: %w", err)
	}

	const q = `
		INSERT INTO plugin_state
			(uuid, plugin_uuid, scope_kind, scope_id, namespace, state_key, value_json, updated_at)
		VALUES
			($1, $2, 'global', NULL, 'default', $3, $4, now())
		ON CONFLICT ON CONSTRAINT plugin_state_unique_entry_idx
		DO UPDATE SET
			value_json = EXCLUDED.value_json,
			updated_at = now()`

	_, err = s.db.ExecContext(ctx, q, uuid.NewString(), pluginUUID, key, raw)
	if err != nil {
		return fmt.Errorf("state set %s/%s: %w", pluginUUID, key, err)
	}
	return nil
}

// Delete removes the key from the store. A no-op if the key does not exist.
func (s *StateStore) Delete(ctx context.Context, pluginUUID, key string) error {
	const q = `
		DELETE FROM plugin_state
		WHERE plugin_uuid = $1
		  AND scope_kind  = 'global'
		  AND scope_id    IS NULL
		  AND namespace   = 'default'
		  AND state_key   = $2`

	_, err := s.db.ExecContext(ctx, q, pluginUUID, key)
	if err != nil {
		return fmt.Errorf("state delete %s/%s: %w", pluginUUID, key, err)
	}
	return nil
}

// List returns all state entries for the given plugin UUID.
func (s *StateStore) List(ctx context.Context, pluginUUID string) ([]StateEntry, error) {
	const q = `
		SELECT uuid, plugin_uuid, scope_kind, scope_id, namespace, state_key, value_json
		FROM plugin_state
		WHERE plugin_uuid = $1
		ORDER BY namespace, state_key`

	var entries []StateEntry
	if err := s.db.SelectContext(ctx, &entries, q, pluginUUID); err != nil {
		return nil, fmt.Errorf("state list %s: %w", pluginUUID, err)
	}
	return entries, nil
}
