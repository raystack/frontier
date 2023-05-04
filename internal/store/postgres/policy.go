package postgres

import (
	"encoding/json"
	"time"

	"github.com/odpf/shield/core/policy"
)

type Policy struct {
	ID           string `db:"id"`
	Role         Role
	RoleID       string `db:"role_id"`
	ResourceID   string `db:"resource_id"`
	Namespace    Namespace
	ResourceType string    `db:"resource_type"`
	UserID       string    `db:"user_id"`
	Metadata     []byte    `db:"metadata"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type PolicyCols struct {
	ID           string `db:"id"`
	RoleID       string `db:"role_id"`
	ResourceType string `db:"resource_type"`
	ResourceID   string `db:"resource_id"`
	UserID       string `db:"user_id"`
	Metadata     []byte `db:"metadata"`
}

func (from Policy) transformToPolicy() (policy.Policy, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return policy.Policy{}, err
		}
	}

	return policy.Policy{
		ID:          from.ID,
		RoleID:      from.RoleID,
		UserID:      from.UserID,
		ResourceID:  from.ResourceID,
		NamespaceID: from.ResourceType,
		Metadata:    unmarshalledMetadata,
	}, nil
}
