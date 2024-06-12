package audit

import (
	"fmt"
	"time"

	"github.com/raystack/frontier/internal/bootstrap/schema"
)

var (
	ErrInvalidDetail = fmt.Errorf("invalid audit details")
	ErrInvalidID     = fmt.Errorf("group id is invalid")
)

type Actor struct {
	ID   string
	Type string
	Name string
}

type Target struct {
	ID   string
	Type string
	Name string
}

// Log is a struct that represents an audit log
type Log struct {
	ID     string
	OrgID  string
	Source string
	Action string

	Actor    Actor
	Target   Target
	Metadata map[string]string

	CreatedAt time.Time
}

type EventName string

func (e EventName) String() string {
	return string(e)
}

const (
	UserCreatedEvent        EventName = "app.user.created"
	UserUpdatedEvent        EventName = "app.user.updated"
	UserDeletedEvent        EventName = "app.user.deleted"
	UserListedEvent         EventName = "app.user.listed"
	ServiceUserCreatedEvent EventName = "app.serviceuser.created"
	ServiceUserDeletedEvent EventName = "app.serviceuser.deleted"

	GroupCreatedEvent       EventName = "app.group.created"
	GroupUpdatedEvent       EventName = "app.group.updated"
	GroupDeletedEvent       EventName = "app.group.deleted"
	GroupMemberRemovedEvent EventName = "app.group.members.removed"

	RoleCreatedEvent EventName = "app.role.created"
	RoleUpdatedEvent EventName = "app.role.updated"
	RoleDeletedEvent EventName = "app.role.deleted"

	PermissionCreatedEvent EventName = "app.permission.created"
	PermissionUpdatedEvent EventName = "app.permission.updated"
	PermissionDeletedEvent EventName = "app.permission.deleted"
	PermissionCheckedEvent EventName = "app.permission.checked"

	BillingEntitlementCheckedEvent EventName = "app.billing.entitlement.checked"

	PolicyCreatedEvent EventName = "app.policy.created"
	PolicyDeletedEvent EventName = "app.policy.deleted"

	OrgCreatedEvent       EventName = "app.organization.created"
	OrgUpdatedEvent       EventName = "app.organization.updated"
	OrgDeletedEvent       EventName = "app.organization.deleted"
	OrgMemberCreatedEvent EventName = "app.organization.member.created"
	OrgMemberDeletedEvent EventName = "app.organization.member.deleted"

	ProjectCreatedEvent EventName = "app.project.created"
	ProjectUpdatedEvent EventName = "app.project.updated"
	ProjectDeletedEvent EventName = "app.project.deleted"

	ResourceCreatedEvent EventName = "app.resource.created"
	ResourceUpdatedEvent EventName = "app.resource.updated"
	ResourceDeletedEvent EventName = "app.resource.deleted"
)

func OrgTarget(id string) Target {
	return Target{
		ID:   id,
		Type: schema.OrganizationNamespace,
	}
}

func ProjectTarget(id string) Target {
	return Target{
		ID:   id,
		Type: schema.ProjectNamespace,
	}
}

func UserTarget(id string) Target {
	return Target{
		ID:   id,
		Type: schema.UserPrincipal,
	}
}

func ServiceUserTarget(id string) Target {
	return Target{
		ID:   id,
		Type: schema.ServiceUserPrincipal,
	}
}

func GroupTarget(id string) Target {
	return Target{
		ID:   id,
		Type: schema.GroupPrincipal,
	}
}
