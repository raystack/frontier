package invitation

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound = errors.New("invitation not found")
)

const (
	inviteEmailSubject = "You have been invited to join an organization"
	inviteEmailBody    = `<div>Hi {{.UserID}},</div>
<br>
<p>You have been invited to join an organization: {{.Organization}}. Login to your account to accept the invitation.</p>
<br>
<div>
Thanks,<br>
Team Frontier
</div>`
)

type Invitation struct {
	ID        uuid.UUID
	UserID    string
	OrgID     string
	GroupIDs  []string
	RoleIDs   []string
	Metadata  metadata.Metadata
	CreatedAt time.Time
	ExpiresAt time.Time
}
