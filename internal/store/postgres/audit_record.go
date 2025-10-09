package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"

	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/server/consts"
)

type AuditRecord struct {
	ID               uuid.UUID          `db:"id" goqu:"skipinsert"`
	IdempotencyKey   uuid.NullUUID      `db:"idempotency_key"`
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
	OrganizationID   uuid.UUID          `db:"org_id"`
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
	var idempotencyKey string
	if ar.IdempotencyKey.Valid {
		idempotencyKey = ar.IdempotencyKey.UUID.String()
	}

	return auditrecord.AuditRecord{
		ID:             ar.ID.String(),
		IdempotencyKey: idempotencyKey,
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

	var idempotencyKey uuid.NullUUID
	if record.IdempotencyKey != "" {
		parsedKey, err := uuid.Parse(record.IdempotencyKey)
		if err != nil {
			return AuditRecord{}, errors.Wrap(err, "invalid idempotency key")
		}
		idempotencyKey = uuid.NullUUID{UUID: parsedKey, Valid: true}
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

func extractActorFromContext(ctx context.Context) (string, string, string, map[string]interface{}) {
	var id, actorType, name string
	var actorMetadata map[string]interface{}

	if val := ctx.Value(consts.AuditRecordActorContextKey); val != nil {
		if actorMap, ok := val.(map[string]interface{}); ok {
			if v, ok := actorMap["id"].(string); ok {
				id = v
			}
			if v, ok := actorMap["type"].(string); ok {
				actorType = v
			}
			if v, ok := actorMap["name"].(string); ok {
				name = v
			}
			if v, ok := actorMap["metadata"].(map[string]interface{}); ok {
				actorMetadata = v
			}
		}
	}
	return id, actorType, name, actorMetadata
}

func extractSessionMetadataFromContext(ctx context.Context) map[string]interface{} {
	if val := ctx.Value(consts.SessionContextKey); val != nil {
		if sessionMetadataMap, ok := val.(map[string]interface{}); ok {
			return sessionMetadataMap
		}
	}
	return nil
}

func extractSuperUserFromContext(ctx context.Context) bool {
	if val := ctx.Value(consts.AuthSuperUserContextKey); val != nil {
		if isSuperUser, ok := val.(bool); ok {
			return isSuperUser
		}
	}
	return false
}

type AuditResource struct {
	ID       string
	Type     string
	Name     string
	Metadata metadata.Metadata
}

type AuditTarget struct {
	ID       string
	Type     string
	Name     string
	Metadata metadata.Metadata
}

// BuildAuditRecord creates an AuditRecord from context and event data
func BuildAuditRecord(ctx context.Context, event string, resource AuditResource, target *AuditTarget, orgID string, eventMetadata metadata.Metadata, occurredAt time.Time) AuditRecord {
	actorID, actorType, actorName, actorMetadata := extractActorFromContext(ctx)

	if actorMetadata == nil {
		actorMetadata = make(map[string]interface{})
	}

	if isSuperUser := extractSuperUserFromContext(ctx); isSuperUser {
		actorMetadata[consts.AuditActorSuperUserKey] = true
	}

	if sessionMetadata := extractSessionMetadataFromContext(ctx); sessionMetadata != nil {
		actorMetadata[consts.AuditSessionMetadataKey] = sessionMetadata
	}

	var actorUUID uuid.UUID
	if actorID == "" { // cron jobs
		actorUUID = uuid.Nil
		actorType = "system"
		actorName = "system"
	} else {
		actorUUID, _ = uuid.Parse(actorID)
	}
	orgUUID, _ := uuid.Parse(orgID)

	record := AuditRecord{
		Event:            event,
		ActorID:          actorUUID,
		ActorType:        actorType,
		ActorName:        actorName,
		ActorMetadata:    metadataToNullJSONText(actorMetadata),
		ResourceID:       resource.ID,
		ResourceType:     resource.Type,
		ResourceName:     resource.Name,
		ResourceMetadata: metadataToNullJSONText(resource.Metadata),
		OrganizationID:   orgUUID,
		OccurredAt:       occurredAt,
		Metadata:         metadataToNullJSONText(eventMetadata),
	}

	// Set target fields if provided
	if target != nil {
		record.TargetID = toNullString(target.ID)
		record.TargetType = toNullString(target.Type)
		record.TargetName = toNullString(target.Name)
		record.TargetMetadata = metadataToNullJSONText(target.Metadata)
	}

	return record
}

// InsertAuditRecordInTx inserts an audit record within a transaction
func InsertAuditRecordInTx(ctx context.Context, tx *sqlx.Tx, record AuditRecord) error {
	query, params, err := dialect.Insert(TABLE_AUDITRECORDS).
		Rows(record).
		ToSQL()
	if err != nil {
		return errors.Wrap(err, "failed to build audit insert query")
	}

	_, err = tx.ExecContext(ctx, query, params...)
	if err != nil {
		return errors.Wrap(err, "failed to insert audit record")
	}

	return nil
}
