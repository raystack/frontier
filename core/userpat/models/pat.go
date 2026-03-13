package models

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
)

type PAT struct {
	ID         string `rql:"name=id,type=string"`
	UserID     string `rql:"name=user_id,type=string"`
	OrgID      string `rql:"name=org_id,type=string"`
	Title      string `rql:"name=title,type=string"`
	SecretHash string `json:"-"`
	Metadata   metadata.Metadata
	RoleIDs    []string   `json:"role_ids,omitempty"`
	ProjectIDs []string   `json:"project_ids,omitempty"`
	LastUsedAt *time.Time `rql:"name=last_used_at,type=datetime"` // last_used_at can be null
	ExpiresAt  time.Time  `rql:"name=expires_at,type=datetime"`
	CreatedAt  time.Time  `rql:"name=created_at,type=datetime"`
	UpdatedAt  time.Time  `rql:"name=updated_at,type=datetime"`
}

type PATList struct {
	PATs []PAT
	Page utils.Page
}
