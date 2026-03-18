package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"abyss-view/internal/domain"
	"abyss-view/internal/middleware"
)

// encryptAESGCM encrypts plaintext with the given 32-byte AES-256-GCM key.
// Returns base64-encoded(iv || ciphertext || tag).
func encryptAESGCM(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("encrypt: key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("encrypt: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("encrypt: new gcm: %w", err)
	}
	iv := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("encrypt: rand iv: %w", err)
	}
	sealed := gcm.Seal(iv, iv, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// decryptAESGCM decrypts base64-encoded ciphertext produced by encryptAESGCM.
func decryptAESGCM(encoded string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("decrypt: key must be 32 bytes")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decrypt: base64 decode: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("decrypt: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("decrypt: new gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("decrypt: ciphertext too short")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: gcm open: %w", err)
	}
	return string(plaintext), nil
}

// loadEncryptionKey loads the 32-byte AES key from ENCRYPTION_KEY env var.
func loadEncryptionKey() ([]byte, error) {
	raw := os.Getenv("ENCRYPTION_KEY")
	if raw == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY environment variable is not set")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err == nil && len(key) == 32 {
		return key, nil
	}
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	return nil, fmt.Errorf("ENCRYPTION_KEY must be a 32-byte raw string or base64-encoded 32 bytes")
}

// SecretService handles CRUD operations for company secrets (metadata only).
type SecretService struct {
	db *sqlx.DB
}

// NewSecretService creates a new SecretService.
func NewSecretService(db *sqlx.DB) *SecretService {
	return &SecretService{db: db}
}

// CreateSecretInput holds fields for creating a secret.
type CreateSecretInput struct {
	Name        string  `json:"name"`
	Value       string  `json:"value"`
	Provider    string  `json:"provider"`
	ExternalRef *string `json:"externalRef"`
	Description *string `json:"description"`
}

// UpdateSecretInput holds fields for updating secret metadata (no value rotation).
type UpdateSecretInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ExternalRef *string `json:"externalRef"`
}

// List returns all secret metadata for a company. Values are never returned.
func (s *SecretService) List(ctx context.Context, companyUUID string) ([]domain.CompanySecret, error) {
	var items []domain.CompanySecret
	err := s.db.SelectContext(ctx, &items, `
		SELECT id AS uuid, company_id AS company_uuid, name, provider,
		       external_ref, latest_version, description,
		       created_by_agent_id AS created_by_agent_uuid, created_by_user_id,
		       created_at, updated_at
		FROM company_secrets
		WHERE company_id = $1
		ORDER BY created_at DESC
	`, companyUUID)
	if err != nil {
		return nil, fmt.Errorf("secrets.List: %w", err)
	}
	return items, nil
}

// Get returns a single secret's metadata by UUID. Value is never returned.
func (s *SecretService) Get(ctx context.Context, companyUUID, secretUUID string) (*domain.CompanySecret, error) {
	var item domain.CompanySecret
	err := s.db.GetContext(ctx, &item, `
		SELECT id AS uuid, company_id AS company_uuid, name, provider,
		       external_ref, latest_version, description,
		       created_by_agent_id AS created_by_agent_uuid, created_by_user_id,
		       created_at, updated_at
		FROM company_secrets
		WHERE id = $1 AND company_id = $2
	`, secretUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Secret not found")
	}
	if err != nil {
		return nil, fmt.Errorf("secrets.Get: %w", err)
	}
	return &item, nil
}

// Create stores a new encrypted secret and returns its metadata.
func (s *SecretService) Create(ctx context.Context, companyUUID string, input CreateSecretInput) (*domain.CompanySecret, error) {
	if input.Name == "" {
		return nil, middleware.ErrUnprocessable("name is required")
	}
	if input.Value == "" {
		return nil, middleware.ErrUnprocessable("value is required")
	}

	key, err := loadEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("secrets.Create: %w", err)
	}
	encryptedValue, err := encryptAESGCM(input.Value, key)
	if err != nil {
		return nil, fmt.Errorf("secrets.Create: encrypt: %w", err)
	}

	provider := input.Provider
	if provider == "" {
		provider = "local_encrypted"
	}

	secretID := uuid.NewString()
	versionID := uuid.NewString()

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("secrets.Create: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO company_secrets (id, company_id, name, provider, external_ref, latest_version, description)
		VALUES ($1, $2, $3, $4, $5, 1, $6)
	`, secretID, companyUUID, input.Name, provider, input.ExternalRef, input.Description)
	if err != nil {
		return nil, fmt.Errorf("secrets.Create: insert secret: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO company_secret_versions (id, secret_id, version, material, value_sha256)
		VALUES ($1, $2, 1, $3, '')
	`, versionID, secretID, encryptedValue)
	if err != nil {
		return nil, fmt.Errorf("secrets.Create: insert version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("secrets.Create: commit: %w", err)
	}

	return s.Get(ctx, companyUUID, secretID)
}

// Update updates secret metadata (name, description, externalRef only).
func (s *SecretService) Update(ctx context.Context, companyUUID, secretUUID string, input UpdateSecretInput) (*domain.CompanySecret, error) {
	existing, err := s.Get(ctx, companyUUID, secretUUID)
	if err != nil {
		return nil, err
	}

	name := existing.Name
	if input.Name != nil {
		name = *input.Name
	}
	description := existing.Description
	if input.Description != nil {
		description = input.Description
	}
	externalRef := existing.ExternalRef
	if input.ExternalRef != nil {
		externalRef = input.ExternalRef
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE company_secrets
		SET name = $1, description = $2, external_ref = $3, updated_at = now()
		WHERE id = $4 AND company_id = $5
	`, name, description, externalRef, secretUUID, companyUUID)
	if err != nil {
		return nil, fmt.Errorf("secrets.Update: %w", err)
	}

	return s.Get(ctx, companyUUID, secretUUID)
}

// Delete permanently removes a secret and all its versions.
func (s *SecretService) Delete(ctx context.Context, companyUUID, secretUUID string) error {
	_, err := s.Get(ctx, companyUUID, secretUUID)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		DELETE FROM company_secrets WHERE id = $1 AND company_id = $2
	`, secretUUID, companyUUID)
	if err != nil {
		return fmt.Errorf("secrets.Delete: %w", err)
	}
	return nil
}
