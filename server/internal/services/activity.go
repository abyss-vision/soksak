package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"soksak/internal/domain"
)

// ActivityService handles recording and querying activity logs.
type ActivityService struct {
	db *sqlx.DB
}

// NewActivityService creates a new ActivityService.
func NewActivityService(db *sqlx.DB) *ActivityService {
	return &ActivityService{db: db}
}

// RecordActivityInput contains the fields required to record an activity event.
type RecordActivityInput struct {
	ActorType  string          `json:"actorType"  validate:"required,oneof=agent user system"`
	ActorID    string          `json:"actorId"    validate:"required"`
	Action     string          `json:"action"     validate:"required"`
	EntityType string          `json:"entityType" validate:"required"`
	EntityID   string          `json:"entityId"   validate:"required"`
	AgentUUID  *string         `json:"agentUuid"`
	RunUUID    *string         `json:"runUuid"`
	Details    json.RawMessage `json:"details"`
}

// ActivityFilters holds optional filters for listing activity.
type ActivityFilters struct {
	AgentUUID  *string
	EntityType *string
	EntityID   *string
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

// RecordActivity inserts a new activity log entry.
func (s *ActivityService) RecordActivity(ctx context.Context, companyUUID string, input RecordActivityInput) (*domain.ActivityLog, error) {
	id := uuid.New().String()
	details := input.Details
	if details == nil {
		details = json.RawMessage("{}")
	}
	var entry domain.ActivityLog
	err := s.db.GetContext(ctx, &entry, `
		INSERT INTO activity_log (
			uuid, company_uuid, actor_type, actor_id, action,
			entity_type, entity_id, agent_uuid, run_uuid, details
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10
		)
		RETURNING uuid, company_uuid, actor_type, actor_id, action,
		          entity_type, entity_id, agent_uuid, run_uuid, details, created_at
	`,
		id, companyUUID, input.ActorType, input.ActorID, input.Action,
		input.EntityType, input.EntityID, input.AgentUUID, input.RunUUID, details,
	)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListActivity returns activity log entries for a company with optional filters.
func (s *ActivityService) ListActivity(ctx context.Context, companyUUID string, filters ActivityFilters) ([]domain.ActivityLog, error) {
	if filters.Limit <= 0 || filters.Limit > 500 {
		filters.Limit = 50
	}
	var entries []domain.ActivityLog
	err := s.db.SelectContext(ctx, &entries, `
		SELECT uuid, company_uuid, actor_type, actor_id, action,
		       entity_type, entity_id, agent_uuid, run_uuid, details, created_at
		FROM activity_log
		WHERE company_uuid = $1
		  AND ($2::text IS NULL OR agent_uuid::text = $2)
		  AND ($3::text IS NULL OR entity_type = $3)
		  AND ($4::text IS NULL OR entity_id = $4)
		  AND ($5::timestamptz IS NULL OR created_at >= $5)
		  AND ($6::timestamptz IS NULL OR created_at <= $6)
		ORDER BY created_at DESC
		LIMIT $7
		OFFSET $8
	`, companyUUID,
		filters.AgentUUID, filters.EntityType, filters.EntityID,
		filters.From, filters.To,
		filters.Limit, filters.Offset,
	)
	if err != nil {
		return nil, err
	}
	return entries, nil
}
