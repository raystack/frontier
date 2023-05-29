package invitation

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/shield/pkg/metadata"
)

var (
	ErrNotFound = errors.New("invitation not found")
)

const (
	inviteEmailSubject = "You have been invited to join an organization"
	inviteEmailBody    = `Hi {{.UserID}},
You have been invited to join an organization: {{.Organization}}. Login to your account to accept the invitation.

Thanks,
Shield Team`
)

type Invitation struct {
	ID        uuid.UUID
	UserID    string
	OrgID     string
	GroupIDs  []string
	Metadata  metadata.Metadata
	CreatedAt time.Time
	ExpiresAt time.Time
}
