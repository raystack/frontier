package organization

import (
	"time"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/metadata"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	Enabled  State = "enabled"
	Disabled State = "disabled"

	AdminPermission = schema.UpdatePermission
	AdminRelation   = schema.OwnerRelationName
	AdminRole       = schema.RoleOrganizationOwner
	MemberRole      = schema.RoleOrganizationViewer
)

type Organization struct {
	ID       string
	Name     string
	Title    string
	Metadata metadata.Metadata
	State    State
	Avatar   string

	CreatedAt time.Time
	UpdatedAt time.Time

	// BillingID is the identifier of the organization in the billing engine
	BillingID string
}
