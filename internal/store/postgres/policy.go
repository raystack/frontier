package postgres

import (
	"encoding/json"
	"time"

	"github.com/raystack/shield/core/policy"
)

type Policy struct {
	ID            string `db:"id"`
	Role          Role
	RoleID        string    `db:"role_id"`
	ResourceID    string    `db:"resource_id"`
	ResourceType  string    `db:"resource_type"`
	PrincipalID   string    `db:"principal_id"`
	PrincipalType string    `db:"principal_type"`
	Metadata      []byte    `db:"metadata"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type PolicyCols struct {
	ID            string `db:"id"`
	RoleID        string `db:"role_id"`
	ResourceType  string `db:"resource_type"`
	ResourceID    string `db:"resource_id"`
	PrincipalID   string `db:"principal_id"`
	PrincipalType string `db:"principal_type"`
	Metadata      []byte `db:"metadata"`
}

func (from Policy) transformToPolicy() (policy.Policy, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return policy.Policy{}, err
		}
	}

	return policy.Policy{
		ID:            from.ID,
		RoleID:        from.RoleID,
		ResourceID:    from.ResourceID,
		ResourceType:  from.ResourceType,
		PrincipalID:   from.PrincipalID,
		PrincipalType: from.PrincipalType,
		Metadata:      unmarshalledMetadata,
	}, nil
}
