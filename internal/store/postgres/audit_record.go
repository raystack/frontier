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
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
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
	OrganizationName string             `db:"org_name"`
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
		Type:     pkgAuditRecord.EntityType(nullStringToString(targetType)),
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
		Event:          pkgAuditRecord.Event(ar.Event),
		Actor: auditrecord.Actor{
			ID:       ar.ActorID.String(),
			Type:     ar.ActorType,
			Name:     ar.ActorName,
			Metadata: nullJSONTextToMetadata(ar.ActorMetadata),
		},
		Resource: auditrecord.Resource{
			ID:       ar.ResourceID,
			Type:     pkgAuditRecord.EntityType(ar.ResourceType),
			Name:     ar.ResourceName,
			Metadata: nullJSONTextToMetadata(ar.ResourceMetadata),
		},
		Target:     nullStringToTargetPtr(ar.TargetID, ar.TargetType, ar.TargetName, ar.TargetMetadata),
		OccurredAt: ar.OccurredAt,
		OrgID:      ar.OrganizationID.String(),
		OrgName:    ar.OrganizationName,
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
		targetType = record.Target.Type.String()
		targetName = record.Target.Name
		targetMetadata = record.Target.Metadata
	}

	return AuditRecord{
		Event:            record.Event.String(),
		ActorID:          actorID,
		ActorType:        record.Actor.Type,
		ActorName:        record.Actor.Name,
		ActorMetadata:    metadataToNullJSONText(record.Actor.Metadata),
		ResourceID:       record.Resource.ID,
		ResourceType:     record.Resource.Type.String(),
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

// enrichActorFromContext enriches actor from context
func enrichActorFromContext(ctx context.Context, actor *auditrecord.Actor) {
	actorID, actorType, actorName, actorMetadata := extractActorFromContext(ctx)

	// Handle system actor (cron jobs, background tasks with no request context)
	if actorID == "" {
		actor.ID = uuid.Nil.String()
		actor.Type = pkgAuditRecord.SystemActor
		actor.Name = pkgAuditRecord.SystemActor
		return
	}

	actor.ID = actorID
	actor.Type = actorType
	actor.Name = actorName
	actor.Metadata = actorMetadata

	// Add additional enrichments
	if actor.Metadata == nil {
		actor.Metadata = make(map[string]interface{})
	}

	if isSuperUser := extractSuperUserFromContext(ctx); isSuperUser {
		actor.Metadata[consts.AuditActorSuperUserKey] = true
	}

	if sessionMetadata := extractSessionMetadataFromContext(ctx); sessionMetadata != nil {
		actor.Metadata[consts.AuditSessionMetadataKey] = sessionMetadata
	}
}

type AuditResource struct {
	ID       string
	Type     pkgAuditRecord.EntityType
	Name     string
	Metadata metadata.Metadata
}

type AuditTarget struct {
	ID       string
	Type     pkgAuditRecord.EntityType
	Name     string
	Metadata metadata.Metadata
}

// BuildAuditRecord creates an AuditRecord from context and event data
func BuildAuditRecord(ctx context.Context, event pkgAuditRecord.Event, resource AuditResource, target *AuditTarget, orgID string, eventMetadata metadata.Metadata, occurredAt time.Time) AuditRecord {
	// Use enrichActorFromContext to get actor details
	var actor auditrecord.Actor
	enrichActorFromContext(ctx, &actor)

	actorUUID, _ := uuid.Parse(actor.ID)
	orgUUID, _ := uuid.Parse(orgID)

	record := AuditRecord{
		Event:            event.String(),
		ActorID:          actorUUID,
		ActorType:        actor.Type,
		ActorName:        actor.Name,
		ActorMetadata:    metadataToNullJSONText(actor.Metadata),
		ResourceID:       resource.ID,
		ResourceType:     resource.Type.String(),
		ResourceName:     resource.Name,
		ResourceMetadata: metadataToNullJSONText(resource.Metadata),
		OrganizationID:   orgUUID,
		OccurredAt:       occurredAt,
		Metadata:         metadataToNullJSONText(eventMetadata),
	}

	// Set target fields if provided
	if target != nil {
		record.TargetID = toNullString(target.ID)
		record.TargetType = toNullString(target.Type.String())
		record.TargetName = toNullString(target.Name)
		record.TargetMetadata = metadataToNullJSONText(target.Metadata)
	}

	return record
}

// InsertAuditRecordInTx inserts an audit record within a transaction
func InsertAuditRecordInTx(ctx context.Context, tx *sqlx.Tx, record AuditRecord) error {
	// Enrich the organization name from DB
	if record.OrganizationID != uuid.Nil {
		var orgName string
		query, params, err := buildOrgNameQuery(record.OrganizationID)
		if err == nil {
			_ = tx.QueryRowContext(ctx, query, params...).Scan(&orgName)
			record.OrganizationName = orgName
		}
	}

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
