package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"abyss-view/internal/domain"
	"abyss-view/internal/middleware"
	"abyss-view/internal/realtime"
)

// ApprovalService handles CRUD and resolution operations for approvals.
type ApprovalService struct {
	db  *sqlx.DB
	hub *realtime.Hub
}

// NewApprovalService creates a new ApprovalService.
func NewApprovalService(db *sqlx.DB, hub *realtime.Hub) *ApprovalService {
	return &ApprovalService{db: db, hub: hub}
}

// CreateApprovalInput contains the fields required to create an approval.
type CreateApprovalInput struct {
	Type                 string          `json:"type"                 validate:"required"`
	Payload              json.RawMessage `json:"payload"              validate:"required"`
	RequestedByAgentUUID *string         `json:"requestedByAgentUuid"`
	RequestedByUserID    *string         `json:"requestedByUserId"`
}

// ResolveApprovalInput contains optional fields for approve/reject decisions.
type ResolveApprovalInput struct {
	DecisionNote    *string `json:"decisionNote"`
	DecidedByUserID *string `json:"decidedByUserId"`
}

const approvalColumns = `
	uuid, company_uuid, type, requested_by_agent_uuid, requested_by_user_id,
	status, payload, decision_note, decided_by_user_id, decided_at,
	created_at, updated_at`

// List returns all approvals for a company, optionally filtered by status.
func (s *ApprovalService) List(ctx context.Context, companyUUID string, status string) ([]domain.Approval, error) {
	var approvals []domain.Approval
	var err error
	if status != "" {
		err = s.db.SelectContext(ctx, &approvals, `
			SELECT `+approvalColumns+`
			FROM approvals
			WHERE company_uuid = $1
			  AND status       = $2
			ORDER BY created_at DESC
		`, companyUUID, status)
	} else {
		err = s.db.SelectContext(ctx, &approvals, `
			SELECT `+approvalColumns+`
			FROM approvals
			WHERE company_uuid = $1
			ORDER BY created_at DESC
		`, companyUUID)
	}
	if err != nil {
		return nil, err
	}
	return approvals, nil
}

// Get returns a single approval by UUID within a company.
func (s *ApprovalService) Get(ctx context.Context, companyUUID, approvalUUID string) (*domain.Approval, error) {
	var approval domain.Approval
	err := s.db.GetContext(ctx, &approval, `
		SELECT `+approvalColumns+`
		FROM approvals
		WHERE uuid         = $1
		  AND company_uuid = $2
	`, approvalUUID, companyUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, middleware.ErrNotFound("Approval not found")
	}
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

// Create creates a new approval within a company.
func (s *ApprovalService) Create(ctx context.Context, companyUUID string, input CreateApprovalInput) (*domain.Approval, error) {
	id := uuid.New().String()

	payload := input.Payload
	if payload == nil {
		payload = json.RawMessage("{}")
	}

	var approval domain.Approval
	err := s.db.GetContext(ctx, &approval, `
		INSERT INTO approvals (
			uuid, company_uuid, type, payload,
			requested_by_agent_uuid, requested_by_user_id,
			status
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			'pending'
		)
		RETURNING `+approvalColumns,
		id, companyUUID, input.Type, payload,
		input.RequestedByAgentUUID, input.RequestedByUserID,
	)
	if err != nil {
		return nil, err
	}

	s.broadcastApproval(realtime.EventApprovalCreated, &approval)
	return &approval, nil
}

// Resolve sets the approval status to approved or rejected.
func (s *ApprovalService) Resolve(ctx context.Context, companyUUID, approvalUUID, targetStatus string, input ResolveApprovalInput) (*domain.Approval, error) {
	existing, err := s.Get(ctx, companyUUID, approvalUUID)
	if err != nil {
		return nil, err
	}

	if existing.Status != "pending" && existing.Status != "revision_requested" {
		if existing.Status == targetStatus {
			return existing, nil
		}
		return nil, &middleware.AppError{
			Code:    "UNPROCESSABLE",
			Status:  422,
			Message: "Only pending or revision_requested approvals can be resolved",
		}
	}

	decidedByUserID := ""
	if input.DecidedByUserID != nil {
		decidedByUserID = *input.DecidedByUserID
	}

	var approval domain.Approval
	err = s.db.GetContext(ctx, &approval, `
		UPDATE approvals
		SET status           = $1,
		    decided_by_user_id = $2,
		    decision_note    = $3,
		    decided_at       = now(),
		    updated_at       = now()
		WHERE uuid         = $4
		  AND company_uuid = $5
		  AND status       IN ('pending', 'revision_requested')
		RETURNING `+approvalColumns,
		targetStatus, decidedByUserID, input.DecisionNote,
		approvalUUID, companyUUID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		// Another concurrent request resolved it; return current state.
		return s.Get(ctx, companyUUID, approvalUUID)
	}
	if err != nil {
		return nil, err
	}

	s.broadcastApproval(realtime.EventApprovalDecided, &approval)
	return &approval, nil
}

// broadcastApproval sends an approval event to all company WebSocket clients.
func (s *ApprovalService) broadcastApproval(eventType string, approval *domain.Approval) {
	if s.hub == nil {
		return
	}
	payload, err := json.Marshal(approval)
	if err != nil {
		return
	}
	s.hub.PublishToCompany(approval.CompanyUUID, realtime.WebSocketMessage{
		Type:      eventType,
		CompanyID: approval.CompanyUUID,
		Payload:   json.RawMessage(payload),
	})
}
