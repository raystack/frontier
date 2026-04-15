package models

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
)

// PATScope pairs a role with its resource type and optional resource IDs.
type PATScope struct {
	RoleID       string   // role UUID
	ResourceType string   // schema.OrganizationNamespace ("app/organization") or schema.ProjectNamespace ("app/project")
	ResourceIDs  []string // specific resource UUIDs; empty = all resources in scope
}

type PAT struct {
	ID            string `rql:"name=id,type=string"`
	UserID        string `rql:"name=user_id,type=string"`
	OrgID         string `rql:"name=org_id,type=string"`
	Title         string `rql:"name=title,type=string"`
	SecretHash    string `json:"-"`
	Metadata      metadata.Metadata
	Scopes        []PATScope
	LastUsedAt    *time.Time `rql:"name=last_used_at,type=datetime"`   // last_used_at can be null
	RegeneratedAt *time.Time `rql:"name=regenerated_at,type=datetime"` // regenerated_at can be null
	ExpiresAt     time.Time  `rql:"name=expires_at,type=datetime"`
	CreatedAt     time.Time  `rql:"name=created_at,type=datetime"`
	UpdatedAt     time.Time  `rql:"name=updated_at,type=datetime"`
}

type PATList struct {
	PATs []PAT
	Page utils.Page
}
