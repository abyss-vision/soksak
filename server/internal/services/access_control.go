package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/middleware"
)

// PermissionGrant represents a row in principal_permission_grants.
type PermissionGrant struct {
	UUID            string    `db:"uuid"              json:"uuid"`
	CompanyUUID     string    `db:"company_uuid"      json:"companyUuid"`
	PrincipalType   string    `db:"principal_type"    json:"principalType"`
	PrincipalID     string    `db:"principal_id"      json:"principalId"`
	PermissionKey   string    `db:"permission_key"    json:"permissionKey"`
	GrantedByUserID *string   `db:"granted_by_user_id" json:"grantedByUserId"`
	CreatedAt       time.Time `db:"created_at"        json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at"        json:"updatedAt"`
}

// GrantPermissionInput contains the fields to create a new permission grant.
type GrantPermissionInput struct {
	PrincipalType   string  `json:"principalType"`
	PrincipalID     string  `json:"principalId"`
	PermissionKey   string  `json:"permissionKey"`
	GrantedByUserID *string `json:"grantedByUserId,omitempty"`
}

// AccessControlService manages permission grants for a company.
type AccessControlService struct {
	db *sqlx.DB
}

// NewAccessControlService creates a new AccessControlService.
func NewAccessControlService(db *sqlx.DB) *AccessControlService {
	return &AccessControlService{db: db}
}

// ListGrants returns all permission grants for a company.
func (s *AccessControlService) ListGrants(ctx context.Context, companyUUID string) ([]PermissionGrant, error) {
	var grants []PermissionGrant
	err := s.db.SelectContext(ctx, &grants, `
		SELECT uuid, company_uuid, principal_type, principal_id,
		       permission_key, granted_by_user_id, created_at, updated_at
		FROM principal_permission_grants
		WHERE company_uuid = $1
		ORDER BY created_at ASC
	`, companyUUID)
	if err != nil {
		return nil, err
	}
	return grants, nil
}

// GrantPermission creates a new permission grant (idempotent on unique key).
func (s *AccessControlService) GrantPermission(ctx context.Context, companyUUID string, input GrantPermissionInput) (*PermissionGrant, error) {
	if input.PrincipalType == "" || input.PrincipalID == "" || input.PermissionKey == "" {
		return nil, middleware.ErrUnprocessable("principalType, principalId, and permissionKey are required")
	}

	id := uuid.New().String()
	var grant PermissionGrant
	err := s.db.GetContext(ctx, &grant, `
		INSERT INTO principal_permission_grants
		    (uuid, company_uuid, principal_type, principal_id, permission_key, granted_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (company_uuid, principal_type, principal_id, permission_key) DO UPDATE
		    SET updated_at = now()
		RETURNING uuid, company_uuid, principal_type, principal_id,
		          permission_key, granted_by_user_id, created_at, updated_at
	`, id, companyUUID, input.PrincipalType, input.PrincipalID, input.PermissionKey, input.GrantedByUserID)
	if err != nil {
		return nil, err
	}
	return &grant, nil
}

// RevokePermission deletes a permission grant by its UUID.
func (s *AccessControlService) RevokePermission(ctx context.Context, companyUUID, grantUUID string) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM principal_permission_grants
		WHERE uuid         = $1
		  AND company_uuid = $2
	`, grantUUID, companyUUID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return middleware.ErrNotFound("Permission grant not found")
	}
	return nil
}

// CheckPermission returns whether a principal holds a given permission key.
func (s *AccessControlService) CheckPermission(ctx context.Context, companyUUID, principalType, principalID, permissionKey string) (bool, error) {
	var grant PermissionGrant
	err := s.db.GetContext(ctx, &grant, `
		SELECT uuid, company_uuid, principal_type, principal_id,
		       permission_key, granted_by_user_id, created_at, updated_at
		FROM principal_permission_grants
		WHERE company_uuid    = $1
		  AND principal_type  = $2
		  AND principal_id    = $3
		  AND permission_key  = $4
	`, companyUUID, principalType, principalID, permissionKey)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
