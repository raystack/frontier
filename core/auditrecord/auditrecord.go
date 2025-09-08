package auditrecord

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	systemActor = "system"
)

type AuditRecord struct {
	ID             string
	Event          string
	Actor          Actor
	Resource       Resource
	Target         *Target
	OccurredAt     time.Time
	OrgID          string
	RequestID      *string
	CreatedAt      time.Time
	Metadata       metadata.Metadata
	IdempotencyKey string
}

type Actor struct {
	ID       string
	Type     string
	Name     string
	Metadata metadata.Metadata
}

type Resource struct {
	ID       string
	Type     string
	Name     string
	Metadata metadata.Metadata
}

type Target struct {
	ID       string
	Type     string
	Name     string
	Metadata metadata.Metadata
}
