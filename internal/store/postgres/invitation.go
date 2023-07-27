package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/invitation"
)

type Invitation struct {
	ID        uuid.UUID `db:"id"`
	UserID    string    `db:"user_id"`
	OrgID     string    `db:"org_id"`
	Metadata  []byte    `db:"metadata"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (from Invitation) transformToInvitation() (invitation.Invitation, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return invitation.Invitation{}, fmt.Errorf("failed to unmarshal metadata of invitation: %w", err)
		}
	}
	var groupIDs []string
	if val, ok := unmarshalledMetadata["group_ids"]; ok && (val != nil) {
		for _, groupIDRaw := range val.([]interface{}) {
			groupIDs = append(groupIDs, groupIDRaw.(string))
		}
	}

	return invitation.Invitation{
		ID:        from.ID,
		UserID:    from.UserID,
		OrgID:     from.OrgID,
		GroupIDs:  groupIDs,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		ExpiresAt: from.ExpiresAt,
	}, nil
}
