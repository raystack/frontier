package auditrecord

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
)

var (
	systemActor = "system"
)

type AuditRecord struct {
	ID             string            `json:"id,omitempty"`
	Event          string            `json:"event"`
	Actor          Actor             `json:"actor"`
	Resource       Resource          `json:"resource"`
	Target         *Target           `json:"target"`
	OccurredAt     time.Time         `json:"occurred_at"`
	OrgID          string            `json:"organization_id"`
	RequestID      *string           `json:"request_id"`
	CreatedAt      time.Time         `json:"created_at,omitempty"`
	Metadata       metadata.Metadata `json:"metadata"`
	IdempotencyKey string            `json:"idempotency_key"`
}

type Actor struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Metadata metadata.Metadata `json:"metadata"`
}

type Resource struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Metadata metadata.Metadata `json:"metadata"`
}

type Target struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Metadata metadata.Metadata `json:"metadata"`
}

type AuditRecordsList struct {
	AuditRecords []AuditRecord
	Group        *utils.Group
	Page         utils.Page
}

// AuditRecordRQLSchema is the schema for audit record RQL queries. This is a flattened version of the AuditRecord struct.
// This is needed because the RQL parser does not support nested structs.
type AuditRecordRQLSchema struct {
	ID             string    `rql:"name=id,type=string"`
	Event          string    `rql:"name=event,type=string"`
	ActorID        string    `rql:"name=actor_id,type=string"`
	ActorType      string    `rql:"name=actor_type,type=string"`
	ActorName      string    `rql:"name=actor_name,type=string"`
	ResourceID     string    `rql:"name=resource_id,type=string"`
	ResourceType   string    `rql:"name=resource_type,type=string"`
	ResourceName   string    `rql:"name=resource_name,type=string"`
	TargetID       string    `rql:"name=target_id,type=string"`
	TargetType     string    `rql:"name=target_type,type=string"`
	TargetName     string    `rql:"name=target_name,type=string"`
	OccurredAt     time.Time `rql:"name=occurred_at,type=datetime"`
	OrgID          string    `rql:"name=organization_id,type=string"`
	RequestID      string    `rql:"name=request_id,type=string"`
	CreatedAt      time.Time `rql:"name=created_at,type=datetime"`
	IdempotencyKey string    `rql:"name=idempotency_key,type=string"`
}
