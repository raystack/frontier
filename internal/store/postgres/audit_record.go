package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"

	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/pkg/metadata"
)

type AuditRecord struct {
	ID               uuid.UUID          `db:"id" goqu:"skipinsert"`
	IdempotencyKey   uuid.UUID          `db:"idempotency_key"`
	Event            string             `db:"event"`
	ActorID          uuid.UUID          `db:"actor_id"`
	ActorType        string             `db:"actor_type"`
	ActorName        string             `db:"actor_name"`
	ActorMetadata    types.NullJSONText `db:"actor_metadata"`
	ResourceID       string             `db:"resource_id"`
	ResourceType     string             `db:"resource_type"`
	ResourceName     string             `db:"resource_name"`
	ResourceMetadata types.NullJSONText `db:"resource_metadata"`
	TargetID         sql.NullString     `db:"target_id"`
	TargetType       sql.NullString     `db:"target_type"`
	TargetName       sql.NullString     `db:"target_name"`
	TargetMetadata   types.NullJSONText `db:"target_metadata"`
	OrganizationID   uuid.UUID          `db:"organization_id"`
	RequestID        sql.NullString     `db:"request_id"`
	OccurredAt       time.Time          `db:"occurred_at"`
	CreatedAt        time.Time          `db:"created_at" goqu:"skipinsert"`
	DeletedAt        sql.NullTime       `db:"deleted_at"`
	Metadata         types.NullJSONText `db:"metadata"`
}

func nullStringToTargetPtr(targetID, targetType, targetName sql.NullString, targetMetadata types.NullJSONText) *auditrecord.Target {
	// Only create Target if at least one field is valid
	if !targetID.Valid && !targetType.Valid && !targetName.Valid && !targetMetadata.Valid {
		return nil
	}

	return &auditrecord.Target{
		ID:       nullStringToString(targetID),
		Type:     nullStringToString(targetType),
		Name:     nullStringToString(targetName),
		Metadata: nullJSONTextToMetadata(targetMetadata),
	}
}

// transformToDomain converts AuditRecord model to domain model
func (ar *AuditRecord) transformToDomain() (auditrecord.AuditRecord, error) {
	return auditrecord.AuditRecord{
		ID:             ar.ID.String(),
		IdempotencyKey: ar.IdempotencyKey.String(),
		Event:          ar.Event,
		Actor: auditrecord.Actor{
			ID:       ar.ActorID.String(),
			Type:     ar.ActorType,
			Name:     ar.ActorName,
			Metadata: nullJSONTextToMetadata(ar.ActorMetadata),
		},
		Resource: auditrecord.Resource{
			ID:       ar.ResourceID,
			Type:     ar.ResourceType,
			Name:     ar.ResourceName,
			Metadata: nullJSONTextToMetadata(ar.ResourceMetadata),
		},
		Target:     nullStringToTargetPtr(ar.TargetID, ar.TargetType, ar.TargetName, ar.TargetMetadata),
		OccurredAt: ar.OccurredAt,
		OrgID:      ar.OrganizationID.String(),
		RequestID:  nullStringToPtr(ar.RequestID),
		CreatedAt:  ar.CreatedAt,
		Metadata:   nullJSONTextToMetadata(ar.Metadata),
	}, nil
}

// transformFromDomain converts auditrecord.AuditRecord domain model to a database model
func transformFromDomain(record auditrecord.AuditRecord) (AuditRecord, error) {
	actorID, err := uuid.Parse(record.Actor.ID)
	if err != nil {
		return AuditRecord{}, errors.Wrap(err, "invalid actor ID")
	}

	orgID, err := uuid.Parse(record.OrgID)
	if err != nil {
		return AuditRecord{}, errors.Wrap(err, "invalid organization ID")
	}

	idempotencyKey, err := uuid.Parse(record.IdempotencyKey)
	if err != nil {
		return AuditRecord{}, errors.Wrap(err, "invalid idempotency key")
	}

	var requestID string
	if record.RequestID != nil {
		requestID = *record.RequestID
	}

	var targetID, targetType, targetName string
	var targetMetadata metadata.Metadata
	if record.Target != nil {
		targetID = record.Target.ID
		targetType = record.Target.Type
		targetName = record.Target.Name
		targetMetadata = record.Target.Metadata
	}

	return AuditRecord{
		Event:            record.Event,
		ActorID:          actorID,
		ActorType:        record.Actor.Type,
		ActorName:        record.Actor.Name,
		ActorMetadata:    metadataToNullJSONText(record.Actor.Metadata),
		ResourceID:       record.Resource.ID,
		ResourceType:     record.Resource.Type,
		ResourceName:     record.Resource.Name,
		ResourceMetadata: metadataToNullJSONText(record.Resource.Metadata),
		TargetID:         toNullString(targetID),
		TargetType:       toNullString(targetType),
		TargetName:       toNullString(targetName),
		TargetMetadata:   metadataToNullJSONText(targetMetadata),
		OrganizationID:   orgID,
		RequestID:        toNullString(requestID),
		OccurredAt:       record.OccurredAt,
		CreatedAt:        record.CreatedAt,
		Metadata:         metadataToNullJSONText(record.Metadata),
		IdempotencyKey:   idempotencyKey,
	}, nil
}
