package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/middleware"
)

// BoardSettings represents per-company board display preferences.
type BoardSettings struct {
	UUID          string          `db:"uuid"            json:"uuid"`
	CompanyUUID   string          `db:"company_uuid"    json:"companyUuid"`
	ColumnOrder   json.RawMessage `db:"column_order"    json:"columnOrder"`
	HiddenColumns json.RawMessage `db:"hidden_columns"  json:"hiddenColumns"`
	SwimLaneField *string         `db:"swim_lane_field" json:"swimLaneField"`
	CreatedAt     time.Time       `db:"created_at"      json:"createdAt"`
	UpdatedAt     time.Time       `db:"updated_at"      json:"updatedAt"`
}

// boardSettingsPatch is the subset of fields that can be PATCHed.
type boardSettingsPatch struct {
	ColumnOrder   *json.RawMessage `json:"columnOrder"`
	HiddenColumns *json.RawMessage `json:"hiddenColumns"`
	SwimLaneField *string          `json:"swimLaneField"`
}

// BoardSettingsRoutes returns a chi.Router for GET/PATCH board-settings.
// Must be mounted under a route that already provides {companyUuid}.
func BoardSettingsRoutes(db *sqlx.DB) chi.Router {
	r := chi.NewRouter()
	r.Get("/", getBoardSettings(db))
	r.Patch("/", patchBoardSettings(db))
	return r
}

func getBoardSettings(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		settings, err := getOrCreateBoardSettings(r.Context(), db, companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get board settings"))
			return
		}
		writeJSON(w, http.StatusOK, settings)
	}
}

func patchBoardSettings(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")

		var patch boardSettingsPatch
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Invalid request body"))
			return
		}

		existing, err := getOrCreateBoardSettings(r.Context(), db, companyUUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to get board settings"))
			return
		}

		// Apply patch fields when provided.
		columnOrder := existing.ColumnOrder
		if patch.ColumnOrder != nil {
			columnOrder = *patch.ColumnOrder
		}
		hiddenColumns := existing.HiddenColumns
		if patch.HiddenColumns != nil {
			hiddenColumns = *patch.HiddenColumns
		}
		swimLaneField := existing.SwimLaneField
		if patch.SwimLaneField != nil {
			swimLaneField = patch.SwimLaneField
		}

		var updated BoardSettings
		err = db.GetContext(r.Context(), &updated, `
			UPDATE board_settings
			SET column_order    = $1,
			    hidden_columns  = $2,
			    swim_lane_field = $3,
			    updated_at      = now()
			WHERE uuid = $4
			RETURNING uuid, company_uuid, column_order, hidden_columns, swim_lane_field, created_at, updated_at
		`, columnOrder, hiddenColumns, swimLaneField, existing.UUID)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to update board settings"))
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

// getOrCreateBoardSettings returns the board_settings row for the company,
// creating a default row if none exists.
func getOrCreateBoardSettings(ctx context.Context, db *sqlx.DB, companyUUID string) (*BoardSettings, error) {
	var row BoardSettings
	err := db.GetContext(ctx, &row, `
		SELECT uuid, company_uuid, column_order, hidden_columns, swim_lane_field, created_at, updated_at
		FROM board_settings
		WHERE company_uuid = $1
	`, companyUUID)
	if err == nil {
		return &row, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Insert default row.
	id := uuid.New().String()
	err = db.GetContext(ctx, &row, `
		INSERT INTO board_settings (uuid, company_uuid, column_order, hidden_columns, swim_lane_field)
		VALUES ($1, $2, '[]'::jsonb, '[]'::jsonb, NULL)
		ON CONFLICT DO NOTHING
		RETURNING uuid, company_uuid, column_order, hidden_columns, swim_lane_field, created_at, updated_at
	`, id, companyUUID)
	if err != nil {
		// Another request may have inserted first; retry read.
		if err2 := db.GetContext(ctx, &row, `
			SELECT uuid, company_uuid, column_order, hidden_columns, swim_lane_field, created_at, updated_at
			FROM board_settings WHERE company_uuid = $1
		`, companyUUID); err2 != nil {
			return nil, err2
		}
	}
	return &row, nil
}
