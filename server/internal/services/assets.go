package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"abyss-view/internal/domain"
	"abyss-view/internal/middleware"
)

// AssetService handles CRUD operations for company assets.
type AssetService struct {
	db *sqlx.DB
}

// NewAssetService creates a new AssetService.
func NewAssetService(db *sqlx.DB) *AssetService {
	return &AssetService{db: db}
}

// CreateAssetInput holds the fields required to record a new asset.
type CreateAssetInput struct {
	Provider           string  `json:"provider"`
	ObjectKey          string  `json:"objectKey"`
	ContentType        string  `json:"contentType"`
	ByteSize           int     `json:"byteSize"`
	SHA256             string  `json:"sha256"`
	OriginalFilename   *string `json:"originalFilename"`
	CreatedByAgentUUID *string `json:"createdByAgentUuid"`
	CreatedByUserID    *string `json:"createdByUserId"`
}

// Get returns a single asset by UUID within a company.
func (s *AssetService) Get(ctx context.Context, companyUUID, assetUUID string) (*domain.Asset, error) {
	var asset domain.Asset
	err := s.db.GetContext(ctx, &asset, `
		SELECT id AS uuid, company_id AS company_uuid, provider, object_key, content_type,
		       byte_size, sha256, original_filename,
		       created_by_agent_id AS created_by_agent_uuid, created_by_user_id,
		       created_at, updated_at
		FROM assets
		WHERE id = $1 AND company_id = $2
	`, assetUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Asset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("assets.Get: %w", err)
	}
	return &asset, nil
}

// GetByUUID returns an asset by UUID without requiring company scope.
// Used internally for download where the companyUUID is derived from the asset.
func (s *AssetService) GetByUUID(ctx context.Context, assetUUID string) (*domain.Asset, error) {
	var asset domain.Asset
	err := s.db.GetContext(ctx, &asset, `
		SELECT id AS uuid, company_id AS company_uuid, provider, object_key, content_type,
		       byte_size, sha256, original_filename,
		       created_by_agent_id AS created_by_agent_uuid, created_by_user_id,
		       created_at, updated_at
		FROM assets
		WHERE id = $1
	`, assetUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Asset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("assets.GetByUUID: %w", err)
	}
	return &asset, nil
}

// Create records a new uploaded asset in the database.
func (s *AssetService) Create(ctx context.Context, companyUUID string, input CreateAssetInput) (*domain.Asset, error) {
	id := uuid.NewString()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO assets (
			id, company_id, provider, object_key, content_type,
			byte_size, sha256, original_filename,
			created_by_agent_id, created_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		id, companyUUID, input.Provider, input.ObjectKey, input.ContentType,
		input.ByteSize, input.SHA256, input.OriginalFilename,
		input.CreatedByAgentUUID, input.CreatedByUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("assets.Create: %w", err)
	}
	return s.Get(ctx, companyUUID, id)
}

// List returns all assets for a company.
func (s *AssetService) List(ctx context.Context, companyUUID string) ([]domain.Asset, error) {
	var assets []domain.Asset
	err := s.db.SelectContext(ctx, &assets, `
		SELECT id AS uuid, company_id AS company_uuid, provider, object_key, content_type,
		       byte_size, sha256, original_filename,
		       created_by_agent_id AS created_by_agent_uuid, created_by_user_id,
		       created_at, updated_at
		FROM assets
		WHERE company_id = $1
		ORDER BY created_at DESC
	`, companyUUID)
	if err != nil {
		return nil, fmt.Errorf("assets.List: %w", err)
	}
	return assets, nil
}
